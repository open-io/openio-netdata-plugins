package zookeeper

import (
    "strings"
    "oionetdata/netdata"
    "oionetdata/util"
    "net"
)

var	ignore = map[string]bool{
	"zk_version": true,
	"zk_server_state": true,
}

/*
Collect - collect zookeeper metrics
*/
func Collect(addr string, ns string, c chan netdata.Metric) {
    conn, err := net.Dial("tcp", addr)

    defer conn.Close()

    util.RaiseIf(err)

    conn.Write([]byte("mntr\n"))

    buff := make([]byte, 4096)
    n, _ := conn.Read(buff)
    stats := strings.Split(string(buff[:n-1]), "\n")

    for s := range stats {
		kv := strings.Split(stats[s], "\t")
		if _, o := ignore[kv[0]]; !o {
            netdata.Update(kv[0], util.SID(addr, ns), kv[1], c)
		}
	}
}
