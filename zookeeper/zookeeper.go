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
	defer conn.Close()

	conn.Write([]byte("mntr\n"))

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

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return data, nil
}
