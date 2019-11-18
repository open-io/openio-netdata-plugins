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

package redis

import (
	"bufio"
    "io/ioutil"
	"fmt"
	"log"
	"net"
	"reflect"
	"testing"
)

type testServer struct {
    specFile string
}

var expected = map[string]string{
    "connected_slaves":"1",
    "instantaneous_ops_per_sec":"39",
    "is_master":"1",
    "keys":"3",
    "keyspace_hits":"2972",
    "keyspace_misses":"176",
    "mem_fragmentation_ratio":"2.19",
    "rdb_changes_since_last_save":"473",
    "repl_backlog_size":"1048576",
    "total_commands_processed":"6973",
    "total_connections_received":"2411",
    "total_net_input_bytes":"328478",
    "total_net_output_bytes":"1053126",
    "used_memory":"2038864",
    "used_memory_lua":"46080",
    "used_memory_rss":"4464640",
}

func newTestServer(specFile string) *testServer {
	return &testServer{specFile: specFile}
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
	data, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		panic(err)
	}
    if data == "QUIT" {
        return
    }
	if data != "INFO\r\n" {
		log.Fatalf("Unknown command %s", data)
	}
    b, err := ioutil.ReadFile(s.specFile)
    if err != nil {
        fmt.Print(err)
    }
	buf := bufio.NewWriter(conn)
	buf.Write(b)
	buf.Flush()
	conn.Close()
}

func TestRedisCollector(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	redis := newTestServer("./redis.spec.txt")
	go redis.Run(l)

	collector := NewCollector(l.Addr().String() + ":0")
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
