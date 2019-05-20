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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/mock/gomock"
	"oionetdata/mock_s3iface"
	"testing"
)

func TestS3Roundtrip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bucket := "test"
	obj := "test"

	mock := mock_s3iface.NewMockS3API(ctrl)
	mock.EXPECT().CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}).AnyTimes()
	mock.EXPECT().PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader([]byte{}),
		Bucket: aws.String(bucket),
		Key:    aws.String(obj),
	}).AnyTimes()
	mock.EXPECT().GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(obj),
	}).AnyTimes()
	mock.EXPECT().DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(obj),
	}).AnyTimes()
	mock.EXPECT().ListObjects(&s3.ListObjectsInput{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(1000),
	}).AnyTimes()
	mock.EXPECT().DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	}).AnyTimes()

	c := collector{
		config:     aws.Config{},
		bucket:     "test",
		object:     "test",
		data:       []byte{},
		objectTtfb: "test",
		dataTtfb:   []byte{},
		Endpoint:   "test",
		s3c:        &s3c{s3: mock},
	}
	data, err := c.Collect()
	if err != nil {
		t.Error(err)
	}

	testResults := map[string]string{
		"2xx":   "1",
		"4xx":   "0",
		"5xx":   "0",
		"other": "0",
	}

	for _, method := range []string{"get", "put", "del", "rb", "mb"} {
		for k, expectedValue := range testResults {
			key := fmt.Sprintf("response_code_%s_%s", method, k)
			if v, ok := data[key]; !ok {
				t.Fatalf("Key %s should have been collected", key)
			} else {
				if v != expectedValue {
					t.Fatalf("Key %s should have value %s, but has %s", key, expectedValue, v)
				}
			}
		}
	}
}
