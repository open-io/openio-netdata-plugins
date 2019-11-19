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
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

type collector struct {
	addr    string
	cluster string
}

func NewCollector(addr string) *collector {
	return &collector{
		addr:    addr,
	}
}

var whitelist = map[string]bool{
	"used_memory":                 true,
	"used_memory_rss":             true,
	"used_memory_lua":             true,
	"mem_fragmentation_ratio":     true,
	"rdb_changes_since_last_save": true,
	"total_connections_received":  true,
	"total_commands_processed":    true,
	"instantaneous_ops_per_sec":   true,
	"total_net_input_bytes":       true,
	"total_net_output_bytes":      true,
	"keyspace_hits":               true,
	"keyspace_misses":             true,
	"role":                        true,
	"connected_slaves":            true,
	"repl_backlog_size":           true,
	"db0":                         true, // Note: VDO: maybe add support for other db?
}

var keysRegexp = regexp.MustCompile(`keys=(\d+)`)

func (c *collector) Collect() (map[string]string, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write([]byte("INFO\r\nQUIT\r\n"))
	if err != nil {
		return nil, err
	}

	data := map[string]string{}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.Split(line, ":")
		if len(kv) != 2 {
			continue
		}
		if _, ok := whitelist[kv[0]]; ok {
			// Match keys in db entry
			if strings.HasPrefix(kv[0], "db") {
				keys := keysRegexp.FindStringSubmatch(kv[1])
				if len(keys) > 1 {
					data["keys"] = keys[1]
				} else {
					fmt.Fprintln(os.Stderr, "WARN: received unparseable db notation", kv[1])
				}
			// Format role:master or role:slave
			} else if kv[0] == "role" {
				if kv[1] == "master" {
					data["is_master"] = "1"
				} else {
					data["is_master"] = "0"
				}
			} else {
				data[kv[0]] = kv[1]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return data, nil
}
