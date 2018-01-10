package main

import (
	"os"
	"strings"
	"flag"
	"oionetdata/openio"
	"oionetdata/netdata"
	"oionetdata/collector"
)

func main() {
	var interval int64;
	os.Args, interval = collector.ParseInterval(os.Args)
	nsPtr := flag.String("ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	confPtr := flag.String("conf", "/etc/oio/sds.conf.d/", "Path to SDS config")
	flag.Parse()

	var proxyURLs = make(map[string]string)
	var namespaces = strings.Split(*nsPtr, ":")
	for i := range namespaces {
		proxyURLs[namespaces[i]] = openio.ProxyURL(*confPtr, namespaces[i]);
	}

	collector.Run(interval, makeCollect(proxyURLs))
}

func makeCollect(proxyURLs map[string]string,) (collect collector.Collect) {
	return func(c chan netdata.Metric) {
		for ns, proxyURL := range proxyURLs {
			openio.Collect(proxyURL, ns, c);
		}
	}
}
