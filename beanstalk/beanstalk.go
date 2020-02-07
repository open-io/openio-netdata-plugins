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
	"net"
	"strings"
)

type collector struct {
	addr  string
	tubes []string
}

func NewCollector(addr string, tubes []string) *collector {
	return &collector{
		addr:  addr,
		tubes: tubes,
	}
}

func SendCommand(conn net.Conn, cmd string, prefix string, data map[string]string) error {
	if _, err := conn.Write([]byte(cmd + "\r\n")); err != nil {
		return err
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		if line == "NOT_FOUND" {
			break
		}
		kv := strings.Split(line, ": ")
		if len(kv) != 2 {
			continue
		}
		data[prefix+kv[0]] = kv[1]
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (c *collector) Collect() (map[string]string, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	data := map[string]string{}

	if err = SendCommand(conn, "stats", "", data); err != nil {
		return nil, err
	}
	for _, tube := range c.tubes {
		if err = SendCommand(conn, "stats-tube "+tube, "_"+tube+"_", data); err != nil {
			return nil, err
		}
	}

	return data, nil
}
