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
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

const multMs = 10e6
const mpuSize = int64(5 * 1024 * 1024)

type s3c struct {
	s3 s3iface.S3API
}

type collector struct {
	config        aws.Config
	bucket        string
	object        string
	data          []byte
	objectTtfb    string
	dataTtfb      []byte
	Endpoint      string
	s3c           *s3c
	bucketCreated bool
	rtBucket      bool
}

func NewCollector(conf map[string]string, rtBucket bool) *collector {
	for _, key := range []string{"endpoint", "access", "secret", "region", "bucket", "object"} {
		if _, ok := conf[key]; !ok {
			log.Fatalf("ERROR: cannot load S3 roundtrip collector: missing '%s' key from config", key)
		}
	}

	var timeout = 5
	if t, ok := conf["timeout"]; ok {
		timeout, _ = strconv.Atoi(t)
	}

	var fileSize = 5*1024*1024 + 2
	if t, ok := conf["size"]; ok {
		fileSize, _ = strconv.Atoi(t)
	}

	var config = aws.Config{
		Region:           aws.String(conf["region"]),
		Credentials:      credentials.NewStaticCredentials(conf["access"], conf["secret"], ""),
		Endpoint:         aws.String(conf["endpoint"]),
		S3ForcePathStyle: aws.Bool(true),
		MaxRetries:       aws.Int(1),
		HTTPClient:       &http.Client{Timeout: time.Second * time.Duration(timeout)},
	}

	sess, err := session.NewSession(&config)
	if err != nil {
		log.Fatalln(err)
	}

	return &collector{
		config:        config,
		bucket:        conf["bucket"],
		object:        conf["object"],
		data:          make([]byte, fileSize),
		objectTtfb:    conf["object"] + "_ttfb",
		dataTtfb:      make([]byte, 1),
		Endpoint:      conf["endpoint"],
		s3c:           &s3c{s3: s3.New(sess)},
		bucketCreated: false,
		rtBucket:      rtBucket,
	}
}

func (c *collector) Collect() (map[string]string, error) {
	data := make(map[string]string)

	for _, req := range []string{"get", "put", "del", "rb", "mb"} {
		for _, dim := range []string{"2xx", "4xx", "5xx", "other"} {
			data[fmt.Sprintf("response_code_%s_%s", req, dim)] = "0"
		}
	}

	if !c.bucketCreated && !c.rtBucket {
		_, _ = c.s3c.mb(c.bucket)
		c.bucketCreated = true
	}

	var time time.Duration
	var err error

	if c.rtBucket {
		time, err = c.s3c.mb(c.bucket)
		register(&data, "mb", code(err), time)
	}

	time, err = c.s3c.put(c.bucket, c.object, c.data)
	register(&data, "put", code(err), time)

	time, err = c.s3c.get(c.bucket, c.object)
	register(&data, "get", code(err), time)

	timeTTFBPut, err := c.s3c.put(c.bucket, c.objectTtfb, c.dataTtfb)
	if err != nil {
		registerTtfb(&data, "put", timeTTFBPut)
		timeTTFBGet, _ := c.s3c.get(c.bucket, c.objectTtfb)
		registerTtfb(&data, "get", timeTTFBGet)

		_, _ = c.s3c.del(c.bucket, c.objectTtfb)
	}

	time, err = c.s3c.ls(c.bucket, 1000)
	register(&data, "ls", code(err), time)

	time, err = c.s3c.del(c.bucket, c.object)
	register(&data, "del", code(err), time)

	if c.rtBucket {
		time, err = c.s3c.rb(c.bucket)
		register(&data, "rb", code(err), time)
	}

	return data, nil
}

func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
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
	input := s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}
	start := time.Now()
	_, err := s.s3.CreateBucket(&input)
	if aerr, ok := err.(awserr.Error); ok {
		if aerr.Code() != s3.ErrCodeBucketAlreadyExists {
			return time.Since(start), nil
		}
	} else {
		return time.Since(start), err
	}
	return time.Since(start), nil
}

func (s *s3c) rb(bucket string) (time.Duration, error) {
	input := s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	}
	start := time.Now()
	_, err := s.s3.DeleteBucket(&input)
	if err != nil {
		return time.Since(start), err
	}
	return time.Since(start), nil
}

func (s *s3c) ls(bucket string, keys int64) (time.Duration, error) {
	input := s3.ListObjectsInput{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(keys),
	}
	start := time.Now()
	_, err := s.s3.ListObjects(&input)
	if err != nil {
		return time.Since(start), err
	}
	return time.Since(start), nil
}

func (s *s3c) cleanupMPU(bucket, obj string, uploadID *string) {
	_, err := s.s3.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(obj),
		UploadId: uploadID,
	})
	if err != nil {
		log.Println("WARN: failed to abort MPU", uploadID, err)
	}
}

func (s *s3c) put(bucket, obj string, data []byte) (time.Duration, error) {
	var part = int64(1)
	var size = int64(len(data))

	start := time.Now()

	// Upload in MPU when size > MPU_SIZE
	if size > mpuSize {
		var parts = []*s3.CompletedPart{}
		resCreate, err := s.s3.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(obj),
		})
		if err != nil {
			return time.Since(start), err
		}

		for size > 0 {
			uploadedSize := min(mpuSize, size)
			input := &s3.UploadPartInput{
				Body:       bytes.NewReader(data[:uploadedSize]),
				Bucket:     aws.String(bucket),
				Key:        aws.String(obj),
				PartNumber: aws.Int64(part),
				UploadId:   resCreate.UploadId,
			}
			resUpload, err := s.s3.UploadPart(input)
			if err != nil {
				s.cleanupMPU(bucket, obj, resCreate.UploadId)
				return time.Since(start), err
			}

			parts = append(parts, &s3.CompletedPart{
				ETag:       resUpload.ETag,
				PartNumber: aws.Int64(part),
			})
			size = size - uploadedSize
			part++
		}

		_, err = s.s3.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
			Bucket:          aws.String(bucket),
			Key:             aws.String(obj),
			MultipartUpload: &s3.CompletedMultipartUpload{Parts: parts},
			UploadId:        resCreate.UploadId,
		})
		if err != nil {
			s.cleanupMPU(bucket, obj, resCreate.UploadId)
			return time.Since(start), err
		}
		return time.Since(start), nil
	}

	_, err := s.s3.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(data),
		Bucket: aws.String(bucket),
		Key:    aws.String(obj),
	})
	if err != nil {
		return time.Since(start), err
	}
	return time.Since(start), nil
}

func (s *s3c) get(bucket, obj string) (time.Duration, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(obj),
	}
	start := time.Now()
	_, err := s.s3.GetObject(input)
	if err != nil {
		return time.Since(start), err
	}
	return time.Since(start), err
}

func (s *s3c) del(bucket, obj string) (time.Duration, error) {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(obj),
	}
	start := time.Now()
	_, err := s.s3.DeleteObject(input)
	if err != nil {
		return time.Since(start), err
	}
	return time.Since(start), err
}
