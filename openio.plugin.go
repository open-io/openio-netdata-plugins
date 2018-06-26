package main

import (
	"flag"
	"oionetdata/collector"
	"oionetdata/netdata"
	"oionetdata/openio"
	"oionetdata/util"
	"os"
	"strings"
)

func main() {
	var interval int64
	os.Args, interval = collector.ParseInterval(os.Args)
	nsPtr := flag.String("ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	confPtr := flag.String("conf", "/etc/oio/sds.conf.d/", "Path to SDS config")
	remotePtr := flag.Bool("remote", false, "Force remote metric collection")
	flag.Parse()

	util.ForceRemote = *remotePtr
	openio.CollectInterval = int(interval)
	var proxyURLs = make(map[string]string)
	var namespaces = strings.Split(*nsPtr, ":")
	for i := range namespaces {
		proxyURLs[namespaces[i]] = openio.ProxyURL(*confPtr, namespaces[i])
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
