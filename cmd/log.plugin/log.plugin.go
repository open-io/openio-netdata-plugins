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
	"oionetdata/netdata"
	"oionetdata/logger"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}
	var conf string
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&conf, "conf", "/etc/netdata/python.d/web_log.conf", "Path to we blog config file")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		log.Fatalln("ERROR: Log plugin: Could not parse args", err)
	}
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	writer := netdata.NewDefaultWriter()
	worker := netdata.NewWorker(time.Duration(intervalSeconds)*time.Second, writer)

	collector := logger.NewCollector(conf)
	worker.AddCollector(collector)

    family := "log"
    // var httpStatuses = []int{"100","101","102","103",
    //     "200","202","203","204","205","206","207","208",
    //     "300","301","302","303","304","305","307","308",
    //     "400","401","402","403","404","405","406","407","408","409","410","411","412","413","414",
    //     "415","416","417","421","422","423","424","425","426","427","428","429","431","451",
    //     "500","501","502","503","504","505","506","507","508","510"}

	responseCode := netdata.NewChart("log", "response_code", "", "Response code", "ops", family, "")
    // Dimensions added dynamically
	// for _, status := range(httpStatuses) {
	// 		dimension := fmt.Sprintf("response_code_%s_%s", status, dim)
	// 		responseCode.AddDimension(dimension, dimension, netdata.AbsoluteAlgorithm)
	// 	}
	// }
	worker.AddChart(responseCode, collector)

	responseTime := netdata.NewChart("log", "response_time", "", "Response time", "ms", family, "")
    // Dimensions added dynamically
	// for _, req := range []string{"min", "max", "avg"} {
	// 	dimension := fmt.Sprintf("response_time_%s", req)
	// 	responseTime.AddDimension(dimension, dimension, netdata.AbsoluteAlgorithm)
	// }
	worker.AddChart(responseTime, collector)

    bandwidthIn := netdata.NewChart("log", "bandwidth_in", "", "Bandwidth in", "kB", family, "")
    // Dimensions added dynamically
	worker.AddChart(bandwidthIn, collector)

    bandwidthOut := netdata.NewChart("log", "bandwidth_out", "", "Bandwidth out", "kB", family, "")
    // Dimensions added dynamically
	worker.AddChart(bandwidthOut, collector)


	worker.Run()
}
