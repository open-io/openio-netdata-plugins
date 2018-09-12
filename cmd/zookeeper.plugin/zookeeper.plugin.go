package main

import (
	"flag"
	"log"
	"os"
	"time"

	"oionetdata/collector"
	"oionetdata/netdata"
	"oionetdata/openio"
	"oionetdata/zookeeper"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}
	var ns string
	var conf string
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&ns, "ns", "OPENIO", "Namespace")
	fs.StringVar(&conf, "conf", "/etc/oio/sds.conf.d/", "Path to SDS config")
	fs.Parse(os.Args[2:])
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

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
