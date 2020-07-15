// OpenIO netdata collectors
// Copyright (C) 2019 OpenIO SAS
//
// This library is free software; you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3.0 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Lesser General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public
// License along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"log"
	"github.com/open-io/openio-netdata-plugins/collector"
	"github.com/open-io/openio-netdata-plugins/container"
	"github.com/open-io/openio-netdata-plugins/netdata"
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
	var addr string
	var limit int64
	var threshold int64
	var fast bool
	var bucketdb string

	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&ns, "ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	fs.StringVar(&conf, "conf", "/etc/oio/sds/", "Path to SDS config directory")
	fs.StringVar(&addr, "addr", "", "Force redis IP:PORT for each namespace")
	fs.StringVar(&bucketdb, "bucketdb", "", "BucketDB redis IP:PORT for each namespace")
	fs.Int64Var(&limit, "limit", -1, "Amount of processed containers in a single request, -1 for unlimited")
	fs.Int64Var(&threshold, "threshold", 3e5, "Minimal number of objects in container to report it")
	fs.BoolVar(&fast, "fast", false, "Use fast account listing")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		log.Fatalln("ERROR: Container plugin: Could not parse args", err)
	}
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	namespaces := strings.Split(ns, ":")
	redisAddr := strings.Split(addr, ",")
	collector.Run(intervalSeconds, makeCollect(conf, redisAddr, strings.Split(bucketdb, ","), namespaces, limit, threshold, fast))
}

func makeCollect(basePath string, addr, bucketdb, namespaces []string, l int64, t int64, f bool) (collect collector.Collect) {

	return func(c chan netdata.Metric) error {
		errors := make(map[string]error)

		for i, ns := range namespaces {
			redisAddr := ""
			var err error
			if i < len(addr) && addr[i] != "" {
				redisAddr = addr[i]
			} else {
				redisAddr, err = container.RedisAddr(basePath, ns)
				if err != nil {
					return err
				}
			}
			client := redis.NewClient(&redis.Options{Addr: redisAddr})

			var bucketDBClient *redis.Client
			if i < len(bucketdb) && bucketdb[i] != "" {
				bucketDBClient = redis.NewClient(&redis.Options{Addr: bucketdb[i]})
			} else {
				bucketDBClient = client
			}

			errors[ns] = container.Collect(client, bucketDBClient, ns, l, t, f, c)
		}
		for _, err := range errors {
			if err != nil {
				return err
			}
		}
		return nil
	}
}
