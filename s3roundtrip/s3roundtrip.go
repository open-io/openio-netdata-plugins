// OpenIO netdata collectors
// Copyright (C) 2019 OpenIO SAS
//
// This library is free software; you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3.0 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Lesser General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public
// License along with this program. If not, see <http://www.gnu.org/licenses/>.

package s3roundtrip

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

const multMs = 1e6

type s3c struct {
	src       io.ReadCloser
	dst       io.Writer
	s3        s3iface.S3API
	dl        s3manageriface.DownloaderAPI
	ul        s3manageriface.UploaderAPI
	userAgent request.Option
	timeout   time.Duration
}

type Object struct {
	name string
	size int64
}

func newObject(name string, size int64) *Object {
	return &Object{
		name: name,
		size: size,
	}
}

type FakeWriterAt struct {
	w io.Writer
}

func (fw FakeWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	return fw.w.Write(p)
}

type collector struct {
	madeBucket   bool
	doMakeBucket bool
	config       aws.Config
	bucket       string
	requests     []string
	objects      map[string]*Object
	Endpoint     string
	s3c          *s3c
}

func NewCollector(conf map[string]string, requests []string) *collector {
	for _, key := range []string{"endpoint", "access", "secret", "region", "bucket", "object"} {
		if _, ok := conf[key]; !ok {
			log.Fatalf("ERROR: cannot load S3 roundtrip collector: missing '%s' key from config", key)
		}
	}

	var timeout = 15 * time.Second
	if t, ok := conf["timeout"]; ok {
		value, err := strconv.Atoi(t)
		if err != nil {
			log.Fatalln("Invalid value for timeout, need integer", value)
		}
		timeout = time.Duration(value) * time.Second
	}

	var mpuSize = 5 * 1024 * 1024
	if t, ok := conf["mpu_size"]; ok {
		value, _ := strconv.Atoi(t)
		if value < mpuSize {
			log.Fatalln("Won't proceed; configured MPU size is lower that minimum MPU Size")
		}
		mpuSize = value
	}

	fileSize := mpuSize - 1
	if t, ok := conf["size"]; ok {
		fileSize, _ = strconv.Atoi(t)
	}

	doMakeBucket := false
	if v, ok := conf["make_bucket"]; ok {
		doMakeBucket = (v == "true")
	}

	ramBuffers := true
	if v, ok := conf["ram_buffers"]; ok {
		ramBuffers = (v == "false")
	}

	config := aws.Config{
		Region:           aws.String(conf["region"]),
		Credentials:      credentials.NewStaticCredentials(conf["access"], conf["secret"], ""),
		Endpoint:         aws.String(conf["endpoint"]),
		S3ForcePathStyle: aws.Bool(true),
		MaxRetries:       aws.Int(1),
		HTTPClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				ExpectContinueTimeout: 0,
			},
		},
	}

	sess, err := session.NewSession(&config)
	if err != nil {
		log.Fatalln(err)
	}

	s3 := s3.New(sess)

	uaString := "OIO-S3RT"
	if v, ok := conf["user_agent"]; ok {
		uaString = v
	}
	userAgent := request.WithAppendUserAgent(uaString)

	input, output := makeInputOutput(ramBuffers, max(mpuSize, fileSize))

	return &collector{
		madeBucket:   false,
		doMakeBucket: doMakeBucket,
		config:       config,
		requests:     requests,
		bucket:       conf["bucket"],
		objects: map[string]*Object{
			"simple": newObject(conf["object"], int64(fileSize)),
			"ttfb":   newObject(conf["object"]+"_ttfb", int64(1)),
			"mpu":    newObject(conf["object"]+"_mpu", int64(mpuSize+1)),
		},
		Endpoint: conf["endpoint"],
		s3c: &s3c{s3: s3,
			src: input,
			dst: output,
			dl: s3manager.NewDownloaderWithClient(s3, func(d *s3manager.Downloader) {
				d.RequestOptions = append(d.RequestOptions, userAgent)
			}),
			ul: s3manager.NewUploaderWithClient(s3, func(u *s3manager.Uploader) {
				u.PartSize = int64(mpuSize)
				u.RequestOptions = append(u.RequestOptions, userAgent)
			}),
			userAgent: userAgent,
			timeout:   timeout,
		},
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func makeInputOutput(ramBuffers bool, size int) (io.ReadCloser, io.Writer) {
	if ramBuffers {
		buf := bytes.NewBuffer(make([]byte, size))
		return ioutil.NopCloser(buf), buf
	}
	fd, err := os.Open("/dev/zero")
	if err != nil {
		log.Fatalln(err)
	}
	return fd, ioutil.Discard
}

func (c *collector) Cleanup() {
	c.s3c.src.Close()
}

func (c *collector) Collect() (map[string]string, error) {
	data := make(map[string]string)

	for _, req := range c.requests {
		for _, dim := range []string{"2xx", "4xx", "5xx", "other"} {
			data[fmt.Sprintf("response_code_%s_%s", req, dim)] = "0"
		}
	}

	if c.doMakeBucket {
		time, err := c.s3c.mb(c.bucket)
		register(&data, "mb", code(err), time)
	} else if !c.doMakeBucket && !c.madeBucket {
		_, _ = c.s3c.mb(c.bucket)
		c.madeBucket = true
	}

	for type_, obj := range c.objects {
		switch type_ {
		case "ttfb":
			timeTTFBPut, err := c.s3c.put(c.bucket, obj)
			if err == nil {
				registerTtfb(&data, "put", timeTTFBPut)
				timeTTFBGet, _ := c.s3c.get(c.bucket, obj)
				registerTtfb(&data, "get", timeTTFBGet)
				timeTTFBGetCache, _ := c.s3c.get(c.bucket, obj)
				registerTtfb(&data, "get_cache", timeTTFBGetCache)
				_, _ = c.s3c.del(c.bucket, obj)
			}
		default:
			pfx := ""
			if type_ == "mpu" {
				pfx = "mpu_"
			}
			timePut, err := c.s3c.put(c.bucket, obj)
			if err == nil {
				timeGet, err := c.s3c.get(c.bucket, obj)
				register(&data, pfx+"get", code(err), timeGet)
				timeDel, err := c.s3c.del(c.bucket, obj)
				register(&data, pfx+"del", code(err), timeDel)
			}
			register(&data, pfx+"put", code(err), timePut)
		}
	}

	timeLs, err := c.s3c.ls(c.bucket, 1000)
	register(&data, "ls", code(err), timeLs)

	if c.doMakeBucket {
		timeRb, err := c.s3c.rb(c.bucket)
		register(&data, "rb", code(err), timeRb)
	}

	return data, nil
}

func code(err error) string {
	if err != nil {
		if req, ok := err.(awserr.RequestFailure); ok {
			// Error is an AWS request failure
			status := req.StatusCode() / 100
			if status == 3 || status == 1 {
				// Treat redirects and continues as 2xx
				status = 2
			}
			return strconv.Itoa(status) + "xx"
		}
		return "other"
	}
	return "2xx"
}

func register(data *map[string]string, req, code string, d time.Duration) {
	(*data)[fmt.Sprintf("response_time_%s", req)] = strconv.FormatInt(d.Nanoseconds()/multMs, 10)
	(*data)[fmt.Sprintf("response_code_%s_%s", req, code)] = "1"
}

func registerTtfb(data *map[string]string, req string, d time.Duration) {
	(*data)[fmt.Sprintf("ttfb_%s", req)] = strconv.FormatInt(d.Nanoseconds()/multMs, 10)
}

func (s *s3c) mb(bucket string) (time.Duration, error) {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	start := time.Now()
	_, err := s.s3.CreateBucketWithContext(ctx, input, s.userAgent)
	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
		return time.Since(start), nil
	}
	return time.Since(start), nil
}

func (s *s3c) rb(bucket string) (time.Duration, error) {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	start := time.Now()
	_, err := s.s3.DeleteBucketWithContext(ctx, input, s.userAgent)
	return time.Since(start), err
}

func (s *s3c) ls(bucket string, keys int64) (time.Duration, error) {
	input := &s3.ListObjectsInput{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(keys),
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	start := time.Now()
	_, err := s.s3.ListObjectsWithContext(ctx, input, s.userAgent)
	return time.Since(start), err
}

func (s *s3c) get(bucket string, obj *Object) (time.Duration, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(obj.name),
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	start := time.Now()
	_, err := s.dl.DownloadWithContext(ctx, FakeWriterAt{w: s.dst}, input)
	return time.Since(start), err
}

func (s *s3c) put(bucket string, obj *Object) (time.Duration, error) {
	input := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(obj.name),
		Body:   io.LimitReader(s.src, obj.size),
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	start := time.Now()
	_, err := s.ul.UploadWithContext(ctx, input)
	return time.Since(start), err
}

func (s *s3c) del(bucket string, obj *Object) (time.Duration, error) {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(obj.name),
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	start := time.Now()
	_, err := s.s3.DeleteObjectWithContext(ctx, input, s.userAgent)
	return time.Since(start), err
}
