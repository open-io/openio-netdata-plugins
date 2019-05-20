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
	"fmt"
	"log"
	"os"
	"time"

	"oionetdata/collector"
	"oionetdata/netdata"
	"oionetdata/s3roundtrip"
	"oionetdata/util"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}
	var conf string
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&conf, "conf", "/etc/netdata/s3-roundtrip.conf", "Path to roundtrip config file")
	fs.Parse(os.Args[2:])
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	writer := netdata.NewDefaultWriter()
	worker := netdata.NewWorker(time.Duration(intervalSeconds)*time.Second, writer)

	config, err := util.S3RoundtripConfig(conf)
	if err != nil {
		log.Fatalln("Could not parse configuration file", err)
	}

	collector := s3roundtrip.NewCollector(config)
	worker.AddCollector(collector)

	responseCode := netdata.NewChart("roundtrip", "response_code", "", "Response code", "ops", collector.Endpoint, "")
	for _, req := range []string{"get", "put", "del", "rb", "mb", "ls"} {
		for _, dim := range []string{"2xx", "4xx", "5xx", "other"} {
			dimension := fmt.Sprintf("response_code_%s_%s", req, dim)
			responseCode.AddDimension(dimension, dimension, netdata.AbsoluteAlgorithm)
		}
	}
	worker.AddChart(responseCode, collector)

	responseTime := netdata.NewChart("roundtrip", "response_time", "", "Response time", "ms", collector.Endpoint, "")
	for _, req := range []string{"get", "put", "del", "rb", "mb", "ls"} {
		dimension := fmt.Sprintf("response_time_%s", req)
		responseTime.AddDimension(dimension, dimension, netdata.AbsoluteAlgorithm)
	}
	worker.AddChart(responseTime, collector)

	ttfb := netdata.NewChart("roundtrip", "ttfb", "", "Time to first byte", "ms", collector.Endpoint, "")
	ttfb.AddDimension("ttfb_put", "ttfb_put", netdata.AbsoluteAlgorithm)
	ttfb.AddDimension("ttfb_get", "ttfb_get", netdata.AbsoluteAlgorithm)
	worker.AddChart(ttfb, collector)

	worker.Run()
}
