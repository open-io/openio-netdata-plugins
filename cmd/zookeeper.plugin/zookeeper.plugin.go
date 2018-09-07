package main

import (
	"flag"
	"log"
	"oionetdata/collector"
	"oionetdata/netdata"
	"oionetdata/openio"
	"oionetdata/zookeeper"
	"strings"
)

func main() {
	var ns string
	var conf string
	flag.StringVar(&ns, "ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	flag.StringVar(&conf, "conf", "/etc/oio/sds.conf.d/", "Path to SDS config")
	flag.Parse()

	intervalSeconds := collector.ParseIntervalSeconds()

	var zkAddrs = make(map[string]string)
	namespaces := strings.Split(ns, ":")
	for _, name := range namespaces {
		addr, err := openio.ZookeeperAddr(conf, name)
		if err != nil {
			log.Fatalf("Load failure: %v", err)
		}
		zkAddrs[name] = addr
	}
	collector.Run(intervalSeconds, makeCollector(zkAddrs))
}

func makeCollector(zkAddrs map[string]string) (collect collector.Collect) {
	return func(c chan netdata.Metric) error {
		for ns, addr := range zkAddrs {
			err := zookeeper.Collect(addr, ns, c)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
