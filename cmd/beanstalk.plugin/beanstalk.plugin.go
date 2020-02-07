// OpenIO netdata collectors
// Copyright (C) 2020 OpenIO SAS
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
	"oionetdata/beanstalk"
	"oionetdata/collector"
	"oionetdata/netdata"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}
	var targets string
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&targets, "targets", "", "Comma separated list of Redis IP:PORT")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		log.Fatalln("ERROR: Beanstalk plugin: Could not parse args", err)
	}
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	if targets == "" {
		log.Fatalln("ERROR: Beanstalk plugin: missing targets")
	}

	writer := netdata.NewDefaultWriter()
	worker := netdata.NewWorker(time.Duration(intervalSeconds)*time.Second, writer)

	for _, target := range strings.Split(targets, ",") {
		res := strings.Split(target, ":")
		if len(res) < 2 {
			log.Fatalln("Invalid parameter", target, "must be IP:PORT[:tube1][:tube2]...")
		}
		addr := res[0] + ":" + res[1]
		tubes := res[2:]
		collector := beanstalk.NewCollector(addr, tubes)
		worker.AddCollector(collector)
		instance := "beanstalk." + addr + ":global"

		c := netdata.NewChart(instance, "jobs", "", "", "", "general", "beanstalk.job")
		c.AddDimension("current-jobs-urgent", "urgent", netdata.AbsoluteAlgorithm)
		c.AddDimension("current-jobs-ready", "ready", netdata.AbsoluteAlgorithm)
		c.AddDimension("current-jobs-reserved", "reserved", netdata.AbsoluteAlgorithm)
		c.AddDimension("current-jobs-delayed", "delayed", netdata.AbsoluteAlgorithm)
		c.AddDimension("current-jobs-buried", "buried", netdata.AbsoluteAlgorithm)
		c.AddDimension("total-jobs", "total", netdata.IncrementalAlgorithm)
		c.AddDimension("jobs-timeouts", "timeouts", netdata.IncrementalAlgorithm)
		worker.AddChart(c, collector)

		c = netdata.NewChart(instance, "commands", "", "", "", "general", "beanstalk.commands")
		c.AddDimension("cmd-put", "put", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-peek", "peek", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-peek-ready", "peek-ready", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-peek-delayed", "peek-delayed", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-peek-buried", "peek-buried", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-reserve", "reserve", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-use", "use", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-watch", "watch", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-ignore", "ignore", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-delete", "delete", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-release", "release", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-bury", "bury", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-kick", "kick", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-stats", "stats", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-stats-job", "stats-job", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-stats-tube", "stats-tube", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-list-tubes", "list-tubes", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-list-tubes-used", "list-tubes-used", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-list-tubes-watched", "list-tubes-watched", netdata.IncrementalAlgorithm)
		c.AddDimension("cmd-pause-tube", "pause-tube", netdata.IncrementalAlgorithm)
		worker.AddChart(c, collector)

		c = netdata.NewChart(instance, "tubes", "", "", "", "general", "beanstalk.tubes")
		c.AddDimension("current-tubes", "current", netdata.AbsoluteAlgorithm)
		worker.AddChart(c, collector)

		c = netdata.NewChart(instance, "connections", "", "", "", "general", "beanstalk.connections")
		c.AddDimension("current-connections", "open", netdata.AbsoluteAlgorithm)
		c.AddDimension("current-producers", "producers", netdata.AbsoluteAlgorithm)
		c.AddDimension("current-workers", "workers", netdata.AbsoluteAlgorithm)
		c.AddDimension("current-waiting", "waiting", netdata.AbsoluteAlgorithm)
		c.AddDimension("total-connections", "total", netdata.IncrementalAlgorithm)
		worker.AddChart(c, collector)

		c = netdata.NewChart(instance, "binlog", "", "", "", "general", "beanstalk.binlog")
		c.AddDimension("binlog-records-written", "written", netdata.IncrementalAlgorithm)
		c.AddDimension("binlog-records-migrated", "compaction", netdata.IncrementalAlgorithm)
		worker.AddChart(c, collector)

		for _, tube := range tubes {
			instance = "beanstalk." + addr + ":" + tube
			c = netdata.NewChart(instance, "jobs", "", "", "", tube, "beanstalk.job")
			c.AddDimension("_"+tube+"_current-jobs-urgent", "urgent", netdata.AbsoluteAlgorithm)
			c.AddDimension("_"+tube+"_current-jobs-ready", "ready", netdata.AbsoluteAlgorithm)
			c.AddDimension("_"+tube+"_current-jobs-reserved", "reserved", netdata.AbsoluteAlgorithm)
			c.AddDimension("_"+tube+"_current-jobs-delayed", "delayed", netdata.AbsoluteAlgorithm)
			c.AddDimension("_"+tube+"_current-jobs-buried", "buried", netdata.AbsoluteAlgorithm)
			c.AddDimension("_"+tube+"_total-jobs", "total", netdata.IncrementalAlgorithm)
			worker.AddChart(c, collector)

			c = netdata.NewChart(instance, "connections", "", "", "", tube, "beanstalk.connections")
			c.AddDimension("_"+tube+"_current-using", "using", netdata.AbsoluteAlgorithm)
			c.AddDimension("_"+tube+"_current-waiting", "waiting", netdata.AbsoluteAlgorithm)
			c.AddDimension("_"+tube+"_current-watching", "watching", netdata.AbsoluteAlgorithm)
			worker.AddChart(c, collector)
		}
	}

	worker.Run()
}
