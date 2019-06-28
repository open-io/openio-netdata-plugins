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
	"net"
	"strings"
)

type collector struct {
	addr string
}

func NewCollector(addr string) *collector {
	return &collector{
		addr: addr,
	}
}

func (c *collector) Collect() (map[string]string, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write([]byte("mntr\n"))
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.Split(line, "\t")
		if len(kv) != 2 {
			continue
		}
		data[kv[0]] = kv[1]
	}
	conn.Close()

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return data, nil
}
