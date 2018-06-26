package main

import (
	"flag"
	"oionetdata/collector"
	"oionetdata/netdata"
	"oionetdata/openio"
	"oionetdata/zookeeper"
	"os"
	"strings"
)

func main() {
	var interval int64
	os.Args, interval = collector.ParseInterval(os.Args)
	nsPtr := flag.String("ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	confPtr := flag.String("conf", "/etc/oio/sds.conf.d/", "Path to SDS config")
	flag.Parse()

	var zkAddrs = make(map[string]string)
	var namespaces = strings.Split(*nsPtr, ":")
	for i := range namespaces {
		a := openio.ZookeeperAddr(*confPtr, namespaces[i])
		if a != "" {
			zkAddrs[namespaces[i]] = a
		}
	}
	collector.Run(interval, makeCollector(zkAddrs))
}

func makeCollector(zkAddrs map[string]string) (collect collector.Collect) {
	return func(c chan netdata.Metric) error {
		for ns, addr := range zkAddrs {
			zookeeper.Collect(addr, ns, c)
		}
		return nil
	}
}
