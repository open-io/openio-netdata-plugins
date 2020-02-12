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
	"os"
	"time"

	"oionetdata/collector"
	nlog "oionetdata/log"
	"oionetdata/netdata"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}
	var conf string
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&conf, "conf", "", "Path to the log config file")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		log.Fatalln("ERROR: Log plugin: Could not parse args", err)
	}
	if _, err := os.Stat(conf); os.IsNotExist(err) {
		log.Fatalln("ERROR: Log plugin: Could not find config at path: '" + conf + "'")
	}

	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	writer := netdata.NewDefaultWriter()
	worker := netdata.NewWorker(time.Duration(intervalSeconds)*time.Second, writer)

	collector := nlog.NewCollector(conf)
	worker.AddCollector(collector)

	family := "log"

	// Dimensions will be added dynamically
	responseCode := netdata.NewChart("log", "response_code", "", "Response code", "ops", family, "")
	worker.AddChart(responseCode, collector)

	responseTime := netdata.NewChart("log", "response_time", "", "Response time", "ms", family, "")
	worker.AddChart(responseTime, collector)

	bandwidthIn := netdata.NewChart("log", "bandwidth_in", "", "Bandwidth in", "kBps", family, "")
	worker.AddChart(bandwidthIn, collector)

	bandwidthOut := netdata.NewChart("log", "bandwidth_out", "", "Bandwidth out", "kBps", family, "")
	worker.AddChart(bandwidthOut, collector)

	worker.Run()
}
