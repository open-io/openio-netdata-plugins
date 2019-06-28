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
	err := fs.Parse(os.Args[2:])
	if err != nil {
		log.Fatalln("ERROR: OpenIO plugin: Could not parse args", err)
	}
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
