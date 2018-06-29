package main

import (
	"flag"
	"github.com/go-redis/redis"
	"oionetdata/collector"
	"oionetdata/container"
	"oionetdata/netdata"
	"os"
	"strings"
)

func main() {
	var interval int64
	os.Args, interval = collector.ParseInterval(os.Args)
	nsPtr := flag.String("ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	confPtr := flag.String("conf", "/etc/oio/sds/", "Path to SDS config directory")
	limitPtr := flag.Int64("limit", -1, "Amount of processed containers in a single request, -1 for unlimited")
	threshPtr := flag.Int64("threshold", 3e5, "Minimal number of objects in container to report it")
	fastPtr := flag.Bool("fast", false, "Use fast account listing")

	flag.Parse()

	collector.Run(interval, makeCollect(*confPtr, strings.Split(*nsPtr, ":"), *limitPtr, *threshPtr, *fastPtr))
}

func makeCollect(basePath string, namespaces []string, l int64, t int64, f bool) (collect collector.Collect) {
	return func(c chan netdata.Metric) error {
		errors := make(map[string]error)
		for nsi := range namespaces {
			client := redis.NewClient(&redis.Options{Addr: container.RedisAddr(basePath, namespaces[nsi])})
			errors[namespaces[nsi]] = container.Collect(client, namespaces[nsi], l, t, f, c)
		}
		for _, err := range errors {
			if err != nil {
				return err
			}
		}
		return nil
	}
}
