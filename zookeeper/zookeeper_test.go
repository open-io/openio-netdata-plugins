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

package zookeeper

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"reflect"
	"testing"
)

type testServer struct {
	data map[string]string
}

func newTestServer(data map[string]string) *testServer {
	return &testServer{data: data}
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
	if data != "mntr\n" {
		log.Fatalf("Unknown command %s", data)
	}
	buf := bufio.NewWriter(conn)
	for k, v := range s.data {
		_, err := buf.WriteString(fmt.Sprintf("%v\t%v\n", k, v))
		if err != nil {
			log.Fatal(err)
		}
	}
	buf.Flush()
	conn.Close()
}

func TestZKCollector(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	testData := map[string]string{
		"foo": "1",
		"bar": "2",
	}
	zk := newTestServer(testData)
	go zk.Run(l)

	collector := NewCollector(l.Addr().String())
	data, err := collector.Collect()
	if err != nil {
		t.Fatalf("unexpected Collect error: %v", err)
	}

	if !reflect.DeepEqual(data, testData) {
		t.Fatalf("unexpected result got\n%v\nexpected\n%v\n", data, testData)
	}

	l.Close()
	_, err = collector.Collect()
	if err == nil {
		t.Fatalf("expected error")
	}
}
