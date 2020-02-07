// OpenIO netdata collectors
// Copyright (C) 2020 OpenIO SAS
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

package beanstalk

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"reflect"
	"strings"
	"testing"
)

type testServer struct {
}

var expected = map[string]string{
	"_default_cmd-delete":                "0",
	"_default_cmd-pause-tube":            "0",
	"_default_current-jobs-buried":       "0",
	"_default_current-jobs-delayed":      "0",
	"_default_current-jobs-ready":        "0",
	"_default_current-jobs-reserved":     "0",
	"_default_current-jobs-urgent":       "0",
	"_default_current-using":             "3",
	"_default_current-waiting":           "12",
	"_default_current-watching":          "31",
	"_default_name":                      "default",
	"_default_pause":                     "0",
	"_default_pause-time-left":           "0",
	"_default_total-jobs":                "0",
	"_oio-delete_cmd-delete":             "90110",
	"_oio-delete_cmd-pause-tube":         "0",
	"_oio-delete_current-jobs-buried":    "0",
	"_oio-delete_current-jobs-delayed":   "0",
	"_oio-delete_current-jobs-ready":     "0",
	"_oio-delete_current-jobs-reserved":  "0",
	"_oio-delete_current-jobs-urgent":    "0",
	"_oio-delete_current-using":          "2",
	"_oio-delete_current-waiting":        "1",
	"_oio-delete_current-watching":       "1",
	"_oio-delete_name":                   "oio-delete",
	"_oio-delete_pause":                  "0",
	"_oio-delete_pause-time-left":        "0",
	"_oio-delete_total-jobs":             "90110",
	"_oio-rebuild_cmd-delete":            "0",
	"_oio-rebuild_cmd-pause-tube":        "0",
	"_oio-rebuild_current-jobs-buried":   "0",
	"_oio-rebuild_current-jobs-delayed":  "0",
	"_oio-rebuild_current-jobs-ready":    "0",
	"_oio-rebuild_current-jobs-reserved": "0",
	"_oio-rebuild_current-jobs-urgent":   "0",
	"_oio-rebuild_current-using":         "2",
	"_oio-rebuild_current-waiting":       "1",
	"_oio-rebuild_current-watching":      "2",
	"_oio-rebuild_name":                  "oio-rebuild",
	"_oio-rebuild_pause":                 "0",
	"_oio-rebuild_pause-time-left":       "0",
	"_oio-rebuild_total-jobs":            "0",
	"_oio_cmd-delete":                    "640670",
	"_oio_cmd-pause-tube":                "0",
	"_oio_current-jobs-buried":           "0",
	"_oio_current-jobs-delayed":          "0",
	"_oio_current-jobs-ready":            "0",
	"_oio_current-jobs-reserved":         "0",
	"_oio_current-jobs-urgent":           "0",
	"_oio_current-using":                 "24",
	"_oio_current-waiting":               "10",
	"_oio_current-watching":              "10",
	"_oio_name":                          "oio",
	"_oio_pause":                         "0",
	"_oio_pause-time-left":               "0",
	"_oio_total-jobs":                    "640670",
	"binlog-current-index":               "149",
	"binlog-max-size":                    "10240000",
	"binlog-oldest-index":                "149",
	"binlog-records-migrated":            "0",
	"binlog-records-written":             "1457478",
	"cmd-bury":                           "0",
	"cmd-delete":                         "728737",
	"cmd-ignore":                         "0",
	"cmd-kick":                           "0",
	"cmd-list-tube-used":                 "0",
	"cmd-list-tubes":                     "32282",
	"cmd-list-tubes-watched":             "0",
	"cmd-pause-tube":                     "0",
	"cmd-peek":                           "0",
	"cmd-peek-buried":                    "0",
	"cmd-peek-delayed":                   "0",
	"cmd-peek-ready":                     "0",
	"cmd-put":                            "728737",
	"cmd-release":                        "4",
	"cmd-reserve":                        "728752",
	"cmd-reserve-with-timeout":           "173970",
	"cmd-stats":                          "128258",
	"cmd-stats-job":                      "0",
	"cmd-stats-tube":                     "129094",
	"cmd-touch":                          "0",
	"cmd-use":                            "28",
	"cmd-watch":                          "13",
	"current-connections":                "31",
	"current-jobs-buried":                "0",
	"current-jobs-delayed":               "0",
	"current-jobs-ready":                 "0",
	"current-jobs-reserved":              "0",
	"current-jobs-urgent":                "0",
	"current-producers":                  "10",
	"current-tubes":                      "4",
	"current-waiting":                    "12",
	"current-workers":                    "12",
	"hostname":                           "node-1.novalocal",
	"id":                                 "fe6081983c4b33dd",
	"job-timeouts":                       "0",
	"max-job-size":                       "65535",
	"pid":                                "2892",
	"rusage-stime":                       "82.499803",
	"rusage-utime":                       "35.948646",
	"total-connections":                  "122597",
	"total-jobs":                         "728737",
	"uptime":                             "348107",
	"version":                            "1.10",
}

func newTestServer() *testServer {
	return &testServer{}
}

func (s *testServer) Run(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("WARN: %s", err)
			return
		}
		go s.handleConn(conn)
	}
}

func (s *testServer) handleConn(conn net.Conn) {
	defer conn.Close()
	out := bufio.NewWriter(conn)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		b, err := ioutil.ReadFile("./beanstalk.spec." + strings.Replace(line, " ", "_", -1) + ".txt")
		if err != nil {
			fmt.Print(err)
			break
		}
		_, err = out.Write(b)
		if err != nil {
			fmt.Print(err)
			break
		}
		out.Flush()
	}
}

func TestBeanstalkCollector(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	beanstalk := newTestServer()
	go beanstalk.Run(l)

	var tubes = []string{"default", "oio", "oio-delete", "oio-rebuild", "prout"}
	collector := NewCollector(l.Addr().String(), tubes)
	data, err := collector.Collect()
	if err != nil {
		t.Fatalf("unexpected Collect error: %v", err)
	}

	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("unexpected result got\n%v\nexpected\n%v\n", data, expected)
	}

	l.Close()
	_, err = collector.Collect()
	if err == nil {
		t.Fatalf("expected error")
	}
}
