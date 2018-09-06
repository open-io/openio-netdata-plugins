package main

import (
	"flag"
	"oionetdata/collector"
	"oionetdata/container"
	"oionetdata/netdata"
	"strings"

	"github.com/go-redis/redis"
)

func main() {
	var ns string
	var conf string
	var limit int64
	var threshold int64
	var fast bool

	flag.StringVar(&ns, "ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	flag.StringVar(&conf, "conf", "/etc/oio/sds/", "Path to SDS config directory")
	flag.Int64Var(&limit, "limit", -1, "Amount of processed containers in a single request, -1 for unlimited")
	flag.Int64Var(&threshold, "threshold", 3e5, "Minimal number of objects in container to report it")
	flag.BoolVar(&fast, "fast", false, "Use fast account listing")

	flag.Parse()
	intervalSeconds := collector.ParseIntervalSeconds()

	namespaces := strings.Split(ns, ":")
	collector.Run(intervalSeconds, makeCollect(conf, namespaces, limit, threshold, fast))
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
