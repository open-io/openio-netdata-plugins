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

package logger

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

func keyExists(t *testing.T, data map[string]string, key string) {
	if _, ok := data[key]; !ok {
		t.Fatalf("Key %s not found in collected result data", key)
	}
}

func sampleConf() (string, string, string) {
	access1, err := ioutil.TempFile("", "access1*")
	if err != nil {
		log.Fatal(err)
	}
	access2, err := ioutil.TempFile("", "access2*")
	if err != nil {
		log.Fatal(err)
	}

	sampleConf := `
# OpenIO Managed
autodetection_retry: 30
OPENIO-rawx-0:
    name: '.openio.OPENIO.rawx.rawx-0.log.access'
    path: '` + access1.Name() + `'
    custom_log_format:
        pattern: '\S+ \S+ \S+ \S+ \S+ \d+ \d+ \S+ \S+ (?P<address>\S+) \S+ (?P<method>\S+) (?P<code>\d+) (?P<resp_time>\d+) (?P<bytes_sent>\d+) \S+ \S+ (?P<url>.*)'
OPENIO-oioswift-0:
    name: '.openio.OPENIO.oioswift.oioswift-0.log'
    path: '` + access2.Name() + `'
    filter:
        include: '- \S+ \d+.\d+ \d+.\d+ -$'
    custom_log_format:
        pattern: '\S+ \S+ \S+ \S+  \S+ (?P<address>\S+) \S+ (?P<method>\S+) (?P<url>\S+) \S+ (?P<code>\d+) \S+ \S+ \S+ (?P<resp_length>\d+|-) (?P<bytes_sent>\d+|-) \S+ \S+ \S+ (?P<resp_time>\d+\.\d+) - (?P<tag_s3_op>\S+) .*'
`
	return sampleConf, access1.Name(), access2.Name()
}

func writeConf(sampleConf string) string {
	content := []byte(sampleConf)
	tmpfile, err := ioutil.TempFile("", "testconf*")
	if err != nil {
		log.Fatal(err)
	}

	fileName := tmpfile.Name()

	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
	return fileName
}

func logToFile(fileName, content string) {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		log.Fatalln(err)
	}
	// Force context switch for tail
	time.Sleep(1000000 * time.Nanosecond)
}

const test10 = "Feb 03 14:23:27 node1 OIO,OPENIO,rawx,0 194 139967883802368 access INF 10.10.10.11:6200 10.10.10.12:55778 DELETE 204 250 99 302 - txd588612fe9d74b9d9f722-005e382cdf /7E6595EB9E87A4D17DB8D39E19FEE8AD91994C55886C24BCE98A8AE6CE4AEBC5\n"
const test20 = "2020-02-03T14:23:27.686225+00:00 node1 OIO,OPENIO,oioswift,0: info  10.10.10.1 10.10.10.1 03/Feb/2020/14/23/27 DELETE /openio-perf-testa1/%257B%2520p%2520.x2%25204%2520%257D%257D/%257B%2520p%2520.x3%25204%2520%257D%257D/%257B%2520p%2520.x4%252016%2520%257D%257Dfbcab794259dc4d2a4aaea67e6cfe132 HTTP/1.0 204 - aws-sdk-go/1.28.9%20%28go1.13.5%3B%20linux%3B%20amd64%29 - - - - tx03b530b5f54f4e90a8122-005e382cdf - 0.0436 - content_delete 1580739807.642102003 1580739807.685687065 -\n"
const test21 = "2020-02-03T14:23:21.862652+00:00 node1 OIO,OPENIO,oioswift,0: info  10.10.10.1 10.10.10.1 03/Feb/2020/14/23/21 PUT /openio-perf-testa1/%257B%2520p%2520.x2%25204%2520%257D%257D/%257B%2520p%2520.x3%25204%2520%257D%257D/%257B%2520p%2520.x4%252016%2520%257D%257D8edbabc62ade5e32e0ba27db2973d2ab HTTP/1.0 200 - aws-sdk-go/1.28.9%20%28go1.13.5%3B%20linux%3B%20amd64%29 - 10000000 - ede3d3b685b4e137ba4cb2521329a75e tx77bd7991046b4295bc16f-005e382cd9 - 0.0509 - content_create 1580739801.810257912 1580739801.861174107 -\n"

// This doesn't match include; ignore
const test30 = "2020-02-03T14:23:27.686225+00:00 node1 OIO,OPENIO,oioswift,0: info  10.10.10.1 10.10.10.1 03/Feb/2020/14/23/27 DELETE /openio-perf-testa1/%257B%2520p%2520.x2%25204%2520%257D%257D/%257B%2520p%2520.x3%25204%2520%257D%257D/%257B%2520p%2520.x4%252016%2520%257D%257Dfbcab794259dc4d2a4aaea67e6cfe132 HTTP/1.0 204 - aws-sdk-go/1.28.9%20%28go1.13.5%3B%20linux%3B%20amd64%29 - - - - tx03b530b5f54f4e90a8122-005e382cdf - 0.0111 S3 content_delete 1580739807.642102003 1580739807.685687065 -\n"

var keysToCheck = []string{
	"bandwidth_in_DELETE.OPENIO-rawx-0",
	"response_time_min_DELETE.OPENIO-oioswift-0.content_delete",
	"response_time_max_DELETE.OPENIO-oioswift-0.content_delete",
	"response_code_log_ops_DELETE.204.OPENIO-oioswift-0.content_delete",
	"response_code_log_ops_PUT.200.OPENIO-oioswift-0.content_create",
	"response_time_max_DELETE.OPENIO-rawx-0",
	"response_time_avg_DELETE.OPENIO-rawx-0",
	"bandwidth_in_PUT.OPENIO-oioswift-0.content_create",
	"response_time_min_DELETE.OPENIO-rawx-0",
	"response_time_avg_DELETE.OPENIO-oioswift-0.content_delete",
	"bandwidth_out_DELETE.OPENIO-oioswift-0.content_delete",
	"response_time_min_PUT.OPENIO-oioswift-0.content_create",
	"response_time_max_PUT.OPENIO-oioswift-0.content_create",
	"response_time_avg_PUT.OPENIO-oioswift-0.content_create",
	"bandwidth_out_PUT.OPENIO-oioswift-0.content_create",
	"response_code_log_ops_DELETE.204.OPENIO-rawx-0",
	"bandwidth_out_DELETE.OPENIO-rawx-0",
	"bandwidth_in_DELETE.OPENIO-oioswift-0.content_delete",
}

func TestLogCollector(t *testing.T) {
	conf, access1, access2 := sampleConf()

	confFile := writeConf(conf)
	defer os.Remove(confFile)
	defer os.Remove(access1)
	defer os.Remove(access2)
	collector := NewCollector(confFile)
	// Wait for access files to be seeked by tail
	time.Sleep(1000000 * time.Nanosecond)

	_, err := collector.Collect()
	if err != nil {
		log.Fatalln(err)
	}

	logToFile(access1, test10)
	logToFile(access2, test20)
	logToFile(access2, test21)
	logToFile(access2, test30)

	res, err := collector.Collect()
	if err != nil {
		log.Fatalln(err)
	}

	for _, key := range keysToCheck {
		keyExists(t, res, key)
	}
}
