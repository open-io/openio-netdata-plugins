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
	"oionetdata/collector"
	"oionetdata/netdata"
	"oionetdata/redis"
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
	fs.StringVar(&targets, "targets", "", "Comma separated list of Redis IP:PORT:CLUSTER_ID")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		log.Fatalln("ERROR: Redis plugin: Could not parse args", err)
	}
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	if targets == "" {
		log.Fatalln("ERROR: Redis plugin: missing targets")
	}

	writer := netdata.NewDefaultWriter()
	worker := netdata.NewWorker(time.Duration(intervalSeconds)*time.Second, writer)

	for _, addr := range strings.Split(targets, ",") {
		collector := redis.NewCollector(addr)
		worker.AddCollector(collector)
		instance := "redis." + addr // TODO colon?

		keysChart := netdata.NewChart(instance, "keys", "", "Keys", "count", instance, "redis.keys.")
		keysChart.AddDimension(
			fmt.Sprintf("keys"),
			"keys",
			netdata.AbsoluteAlgorithm, //netdata.IncrementalAlgorithm,
		)
		worker.AddChart(keysChart, collector)

		memChart := netdata.NewChart(instance, "memory", "", "Memory", "bytes", instance, "redis.memory")
		for k, v := range map[string]string{
			"used_memory": "total", "used_memory_rss": "rss", "used_memory_lua": "lua"} {
			memChart.AddDimension(k, v, netdata.AbsoluteAlgorithm) //netdata.IncrementalAlgorithm,)
		}
		worker.AddChart(memChart, collector)

		bandwidthChart := netdata.NewChart(instance, "net", "", "Network traffic", "bytes", instance, "redis.net")
		for k, v := range map[string]string{"total_net_input_bytes": "received", "total_net_output_bytes": "sent"} {
			bandwidthChart.AddDimension(k, v, netdata.IncrementalAlgorithm) //netdata.IncrementalAlgorithm,)
		}
		worker.AddChart(bandwidthChart, collector)

		opsChart := netdata.NewChart(instance, "instant", "", "Instantaneous operations", "ops", instance, "redis.ops")
		opsChart.AddDimension("instantaneous_ops_per_sec", "ops", netdata.AbsoluteAlgorithm)
		worker.AddChart(opsChart, collector)

		masterChart := netdata.NewChart(instance, "state", "", "Instance is master", "master", instance, "redis.master")
		masterChart.AddDimension("is_master", "state", netdata.AbsoluteAlgorithm)
		worker.AddChart(masterChart, collector)

		replicaCharts := netdata.NewChart(instance, "replicas", "", "Replicas", "count", instance, "redis.replicas")
		replicaCharts.AddDimension("connected_slaves", "replicas", netdata.AbsoluteAlgorithm)
		worker.AddChart(replicaCharts, collector)

		cacheCharts := netdata.NewChart(instance, "cache", "", "Cache", "ops", instance, "redis.cache")
		cacheCharts.AddDimension("keyspace_hits", "hits", netdata.AbsoluteAlgorithm)
		cacheCharts.AddDimension("keyspace_misses", "misses", netdata.AbsoluteAlgorithm)
		worker.AddChart(cacheCharts, collector)

		backlogCharts := netdata.NewChart(instance, "backlog", "", "Backlog", "bytes", instance, "redis.backlog")
		backlogCharts.AddDimension("repl_backlog_size", "backlog", netdata.AbsoluteAlgorithm)
		worker.AddChart(backlogCharts, collector)

		changesSinceSave := netdata.NewChart(instance, "changes", "", "Changes since last save", "ops", instance, "redis.changes")
		changesSinceSave.AddDimension("rdb_changes_since_last_save", "changes", netdata.AbsoluteAlgorithm)
		worker.AddChart(changesSinceSave, collector)

		connCharts := netdata.NewChart(instance, "connections", "", "Connections", "count", instance, "redis.connections")
		connCharts.AddDimension("total_connections_received", "connections", netdata.AbsoluteAlgorithm)
		worker.AddChart(connCharts, collector)

		commandCharts := netdata.NewChart(instance, "commands", "", "Commands", "count", instance, "redis.commands")
		commandCharts.AddDimension("total_commands_processed", "commands", netdata.AbsoluteAlgorithm)
		worker.AddChart(commandCharts, collector)

		memFragmentCharts := netdata.NewChart(instance, "fragmentation", "", "Memory fragmentation", "ratio", instance, "redis.fragmentation")
		memFragmentCharts.AddDimension("mem_fragmentation_ratio", "fragmentation", netdata.AbsoluteAlgorithm)
		worker.AddChart(memFragmentCharts, collector)
	}

	worker.Run()
}
