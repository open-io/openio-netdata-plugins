package main

import (
	"flag"
	"log"
	"oionetdata/collector"
	"oionetdata/netdata"
	"oionetdata/openio"
	"oionetdata/util"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}

	var ns string
	var conf string
	var remote bool

	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&ns, "ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	fs.StringVar(&conf, "conf", "/etc/oio/sds.conf.d/", "Path to SDS config")
	fs.BoolVar(&remote, "remote", false, "Force remote metric collection")
	fs.Parse(os.Args[2:])
	interval := collector.ParseIntervalSeconds(os.Args[1])

	util.ForceRemote = remote
	openio.CollectInterval = int(interval)
	var proxyURLs = make(map[string]string)
	namespaces := strings.Split(ns, ":")
	for _, name := range namespaces {
		addr, err := openio.ProxyAddr(conf, name)
		if err != nil {
			log.Fatalf("Load failure: %v", err)
		}
		proxyURLs[name] = addr
	}

	collector.Run(interval, makeCollect(proxyURLs))
}

func makeCollect(proxyURLs map[string]string) (collect collector.Collect) {
	return func(c chan netdata.Metric) error {
		for ns, proxyURL := range proxyURLs {
			openio.Collect(proxyURL, ns, c)
		}
		return nil
	}
}
