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

package memcached

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"reflect"
	"testing"
)

type testServer struct {
	specFile string
}

var expected = map[string]string{
	"accepting_conns":               "1",
	"auth_cmds":                     "0",
	"auth_errors":                   "0",
	"bytes":                         "50636054",
	"bytes_read":                    "477545177",
	"bytes_written":                 "505475773",
	"cas_badval":                    "0",
	"cas_hits":                      "0",
	"cas_misses":                    "0",
	"cmd_flush":                     "0",
	"cmd_get":                       "2673422",
	"cmd_set":                       "427765",
	"cmd_touch":                     "0",
	"conn_yields":                   "0",
	"connection_structures":         "59",
	"crawler_items_checked":         "1408456",
	"crawler_reclaimed":             "747",
	"curr_connections":              "58",
	"curr_items":                    "25285",
	"decr_hits":                     "0",
	"decr_misses":                   "0",
	"delete_hits":                   "60238",
	"delete_misses":                 "46367",
	"direct_reclaims":               "105006",
	"evicted_active":                "0",
	"evicted_unfetched":             "105002",
	"evictions":                     "105006",
	"expired_unfetched":             "2432",
	"get_expired":                   "467",
	"get_flushed":                   "0",
	"get_hits":                      "1697871",
	"get_misses":                    "975551",
	"hash_bytes":                    "524288",
	"hash_is_expanding":             "0",
	"hash_power_level":              "16",
	"incr_hits":                     "0",
	"incr_misses":                   "0",
	"libevent":                      "2.0.21-stable",
	"limit_maxbytes":                "67108864",
	"listen_disabled_num":           "0",
	"log_watcher_sent":              "0",
	"log_watcher_skipped":           "0",
	"log_worker_dropped":            "0",
	"log_worker_written":            "0",
	"lru_bumps_dropped":             "0",
	"lru_crawler_running":           "0",
	"lru_crawler_starts":            "19393",
	"lru_maintainer_juggles":        "3691122",
	"lrutail_reflocked":             "21395",
	"malloc_fails":                  "0",
	"max_connections":               "1024",
	"moves_to_cold":                 "228319",
	"moves_to_warm":                 "111595",
	"moves_within_lru":              "269852",
	"pid":                           "126004",
	"pointer_size":                  "64",
	"reclaimed":                     "4399",
	"rejected_connections":          "0",
	"reserved_fds":                  "20",
	"rusage_system":                 "95.562515",
	"rusage_user":                   "54.346651",
	"slab_global_page_pool":         "0",
	"slab_reassign_busy_deletes":    "0",
	"slab_reassign_busy_items":      "0",
	"slab_reassign_chunk_rescues":   "0",
	"slab_reassign_evictions_nomem": "0",
	"slab_reassign_inline_reclaim":  "0",
	"slab_reassign_rescues":         "0",
	"slab_reassign_running":         "0",
	"slabs_moved":                   "0",
	"threads":                       "4",
	"time":                          "1580995111",
	"time_in_listen_disabled_us":    "0",
	"total_connections":             "101",
	"total_items":                   "427765",
	"touch_hits":                    "0",
	"touch_misses":                  "0",
	"uptime":                        "164717",
	"version":                       "1.5.6",
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
	if data != "stats\r\n" {
		log.Fatalf("Unknown command %s", data)
	}
	b, err := ioutil.ReadFile(s.specFile)
	if err != nil {
		fmt.Print(err)
	}
	buf := bufio.NewWriter(conn)
	_, err = buf.Write(b)
	if err != nil {
		fmt.Print(err)
	}
	buf.Flush()
	conn.Close()
}

func TestMemcachedCollector(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	memcached := newTestServer("./memcached.spec.txt")
	go memcached.Run(l)

	collector := NewCollector(l.Addr().String())
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
