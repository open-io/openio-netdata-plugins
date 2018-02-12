package main

import (
    "os"
    "strings"
    "flag"
    "oionetdata/collector"
    "oionetdata/netdata"
    "oionetdata/container"
)

func main() {
	var interval int64
	os.Args, interval = collector.ParseInterval(os.Args)
	nsPtr := flag.String("ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	confPtr := flag.String("conf", "/etc/oio/sds/", "Path to SDS config directory")
    limitPtr := flag.Int64("limit", -1, "Amount of processed containers in a single request, -1 for unlimited")
    threshPtr := flag.Int64("threshold", 3e5, "Minimal number of objects in container to report it")

	flag.Parse()

	collector.Run(interval, makeCollect(*confPtr, strings.Split(*nsPtr, ":"), *limitPtr, *threshPtr))
}

func makeCollect(basePath string, namespaces []string, l int64, t int64) (collect collector.Collect) {
	return func(c chan netdata.Metric) {
		for nsi := range namespaces {
			container.Collect(basePath, namespaces[nsi], l, t, c);
		}
	}
}
