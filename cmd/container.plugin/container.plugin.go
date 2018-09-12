package main

import (
	"flag"
	"log"
	"oionetdata/collector"
	"oionetdata/container"
	"oionetdata/netdata"
	"os"
	"strings"

	"github.com/go-redis/redis"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}

	var ns string
	var conf string
	var limit int64
	var threshold int64
	var fast bool

	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&ns, "ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	fs.StringVar(&conf, "conf", "/etc/oio/sds/", "Path to SDS config directory")
	fs.Int64Var(&limit, "limit", -1, "Amount of processed containers in a single request, -1 for unlimited")
	fs.Int64Var(&threshold, "threshold", 3e5, "Minimal number of objects in container to report it")
	fs.BoolVar(&fast, "fast", false, "Use fast account listing")
	fs.Parse(os.Args[2:])
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	namespaces := strings.Split(ns, ":")
	collector.Run(intervalSeconds, makeCollect(conf, namespaces, limit, threshold, fast))
}

func makeCollect(basePath string, namespaces []string, l int64, t int64, f bool) (collect collector.Collect) {

	return func(c chan netdata.Metric) error {
		errors := make(map[string]error)
		for _, ns := range namespaces {
			redisAddr, err := container.RedisAddr(basePath, ns)
			if err != nil {
				return err
			}
			client := redis.NewClient(&redis.Options{Addr: redisAddr})
			errors[ns] = container.Collect(client, ns, l, t, f, c)
		}
		for _, err := range errors {
			if err != nil {
				return err
			}
		}
		return nil
	}
}
