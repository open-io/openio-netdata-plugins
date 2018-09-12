package zookeeper

import (
	"bufio"
	"fmt"
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
			return
		}
		go s.handleConn(conn)
	}
}

func (s *testServer) handleConn(conn net.Conn) {
	buf := bufio.NewWriter(conn)
	for k, v := range s.data {
		buf.WriteString(fmt.Sprintf("%v\t%v\n", k, v))
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
