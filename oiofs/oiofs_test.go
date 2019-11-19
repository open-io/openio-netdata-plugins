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

package oiofs

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

var testEndpoints = []Endpoint{
	{Path: "/mnt/test", URL: "127.0.0.1:37000"},
	{Path: "/mnt/test2", URL: "127.0.0.1:37001"},
}

type testServer struct {
	specFile string
	addr     string
}

func newTestServer(specFile string, addr string) *testServer {
	return &testServer{specFile: specFile, addr: addr}
}

func (s *testServer) Run() {
	server := http.NewServeMux()
	server.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadFile(s.specFile)
		if err != nil {
			fmt.Print(err)
		}
		_, err = w.Write(b)
		if err != nil {
			fmt.Print(err)
		}
	})
	err := http.ListenAndServe(s.addr, server)
	if err != nil {
		log.Fatal(err)
	}
}

func TestOiofsCollector(t *testing.T) {
	srv := newTestServer("./oiofs.spec.json", "127.0.0.1:37000")
	go srv.Run()
	srv2 := newTestServer("./oiofs.spec.json", "127.0.0.1:37001")
	go srv2.Run()

	tests := []map[string]int64{
		map[string]int64{
			"sds_upload_total_byte":            1234,
			"fuse_write_total_byte":            0,
			"cache_chunk_avg_age_microseconds": 0,
			// Debug options
			"fuse_flush_total_us":   -1,
			"fuse_create_count":     -1,
			"Meta_init_ctx_count":   -1,
			"Meta_SetLink_total_us": -1,
			"sds_StatFs_total_us":   -1,
			"fuse_flush_max_us":     -1,
			"sds_StatFs_avg_us":     -1,
		},
		map[string]int64{
			"sds_upload_total_byte":            1234,
			"fuse_write_total_byte":            0,
			"cache_chunk_avg_age_microseconds": 0,
			// Debug options
			"fuse_flush_total_us":   0,
			"fuse_create_count":     0,
			"Meta_init_ctx_count":   1,
			"Meta_SetLink_total_us": 0,
			"sds_StatFs_total_us":   0,
			"fuse_flush_max_us":     -1,
			"sds_StatFs_avg_us":     -1,
		},
	}

	start := time.Now()

	for {
		now := time.Now()
		elapsed := now.Sub(start)
		if elapsed > 30*time.Second {
			t.Fatal("Could not connect to testserver endpoints")
		}
		errors := 0
		for _, endpoint := range testEndpoints {
			conn, err := net.Dial("tcp", endpoint.URL)
			if err != nil {
				errors++
				break
			} else {
				conn.Close()
			}
		}
		if errors == 0 {
			break
		}
		time.Sleep(time.Second)
	}

	for _, test := range []int{0, 1} {
		func(full int) {
			for _, endpoint := range testEndpoints {
				c := NewCollector(endpoint, full == 1)
				res, err := c.Collect()

				if err != nil {
					t.Fatal(err)
				}

				// Test returned data
				for k, v := range tests[full] {
					if v < 0 {
						if _, ok := res[k]; ok {
							t.Fatalf("Key %s shouldn't have been collected (full: %d)", k, full)
						}
					} else {
						if v2, ok := res[k]; !ok || v2 != strconv.FormatInt(v, 10) {
							t.Fatalf("Key %s not found in collected result data (full: %d)", k, full)
						}
					}
				}
			}
		}(test)
	}
}
