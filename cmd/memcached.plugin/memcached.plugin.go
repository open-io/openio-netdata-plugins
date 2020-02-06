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
	"oionetdata/collector"
	"oionetdata/memcached"
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
		log.Fatalln("ERROR: Memcached plugin: Could not parse args", err)
	}
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	if targets == "" {
		log.Fatalln("ERROR: Memcached plugin: missing targets")
	}

	writer := netdata.NewDefaultWriter()
	worker := netdata.NewWorker(time.Duration(intervalSeconds)*time.Second, writer)

	for _, addr := range strings.Split(targets, ",") {
		res := strings.Split(addr, ":")
		if len(res) != 2 {
			log.Fatalln("Invalid address", addr, "must be IP:PORT")
		}
		collector := memcached.NewCollector(res[0] + ":" + res[1])
		worker.AddCollector(collector)
		instance := "memcached." + addr

		uptimeChart := netdata.NewChart(instance, "uptime", "", "Uptime", "seconds", instance, "memcached.uptime.")
		uptimeChart.AddDimension("uptime", "current", netdata.AbsoluteAlgorithm)
		worker.AddChart(uptimeChart, collector)

		itemsChart := netdata.NewChart(instance, "items", "", "Items", "count", instance, "memcached.items.")
		itemsChart.AddDimension("curr_items", "current", netdata.AbsoluteAlgorithm)
		itemsChart.AddDimension("total_items", "total", netdata.IncrementalAlgorithm)
		worker.AddChart(itemsChart, collector)

		memChart := netdata.NewChart(instance, "memory", "", "Memory", "bytes", instance, "memcached.memory")
		memChart.AddDimension("bytes", "current", netdata.AbsoluteAlgorithm)
		memChart.AddDimension("limit_maxbytes", "max", netdata.AbsoluteAlgorithm)
		worker.AddChart(memChart, collector)

		connectionsChart := netdata.NewChart(instance, "connections", "", "Connections", "count", instance, "memcached.connections")
		connectionsChart.AddDimension("max_connections", "max", netdata.AbsoluteAlgorithm)
		connectionsChart.AddDimension("curr_connections", "current", netdata.AbsoluteAlgorithm)
		connectionsChart.AddDimension("total_connections", "total", netdata.IncrementalAlgorithm)
		connectionsChart.AddDimension("rejected_connections", "rejected", netdata.IncrementalAlgorithm)
		connectionsChart.AddDimension("accepting_conns", "accepting", netdata.AbsoluteAlgorithm)
		connectionsChart.AddDimension("listen_disabled_num", "disabled", netdata.AbsoluteAlgorithm)
		connectionsChart.AddDimension("conn_yields", "yield", netdata.AbsoluteAlgorithm)
		worker.AddChart(connectionsChart, collector)

		reqsChart := netdata.NewChart(instance, "requests", "", "Requests", "requests", instance, "memcached.requests")
		reqsChart.AddDimension("cmd_get", "get", netdata.IncrementalAlgorithm)
		reqsChart.AddDimension("cmd_set", "set", netdata.IncrementalAlgorithm)
		reqsChart.AddDimension("cmd_flush", "flush", netdata.IncrementalAlgorithm)
		reqsChart.AddDimension("cmd_touch", "touch", netdata.IncrementalAlgorithm)
		worker.AddChart(reqsChart, collector)

		getsChart := netdata.NewChart(instance, "get_requests", "", "Get requests", "requests", instance, "memcached.get_requests")
		getsChart.AddDimension("get_hits", "hits", netdata.IncrementalAlgorithm)
		getsChart.AddDimension("get_misses", "misses", netdata.IncrementalAlgorithm)
		getsChart.AddDimension("get_expired", "expired", netdata.IncrementalAlgorithm)
		getsChart.AddDimension("get_flushed", "flushed", netdata.IncrementalAlgorithm)
		worker.AddChart(getsChart, collector)

		deletesChart := netdata.NewChart(instance, "delete_requests", "", "Delete requests", "requests", instance, "memcached.delete_requests")
		deletesChart.AddDimension("delete_hits", "hits", netdata.IncrementalAlgorithm)
		deletesChart.AddDimension("delete_misses", "misses", netdata.IncrementalAlgorithm)
		worker.AddChart(deletesChart, collector)

		incrsChart := netdata.NewChart(instance, "incr_requests", "", "Incr requests", "requests", instance, "memcached.incr_requests")
		incrsChart.AddDimension("incr_hits", "hits", netdata.IncrementalAlgorithm)
		incrsChart.AddDimension("incr_misses", "misses", netdata.IncrementalAlgorithm)
		worker.AddChart(incrsChart, collector)

		decrsChart := netdata.NewChart(instance, "decr_requests", "", "Decr requests", "requests", instance, "memcached.decr_requests")
		decrsChart.AddDimension("decr_hits", "hits", netdata.IncrementalAlgorithm)
		decrsChart.AddDimension("decr_misses", "misses", netdata.IncrementalAlgorithm)
		worker.AddChart(decrsChart, collector)

		cassChart := netdata.NewChart(instance, "cas_requests", "", "CAS requests", "requests", instance, "memcached.cas_requests")
		cassChart.AddDimension("cas_hits", "hits", netdata.IncrementalAlgorithm)
		cassChart.AddDimension("cas_misses", "misses", netdata.IncrementalAlgorithm)
		cassChart.AddDimension("cas_bandval", "badval", netdata.IncrementalAlgorithm)
		worker.AddChart(cassChart, collector)

		touchsChart := netdata.NewChart(instance, "touch_requests", "", "Touch requests", "requests", instance, "memcached.touch_requests")
		touchsChart.AddDimension("touch_hits", "hits", netdata.IncrementalAlgorithm)
		touchsChart.AddDimension("touch_misses", "misses", netdata.IncrementalAlgorithm)
		worker.AddChart(touchsChart, collector)

		authsChart := netdata.NewChart(instance, "auth_requests", "", "Auth requests", "requests", instance, "memcached.auth_requests")
		authsChart.AddDimension("auth_cmds", "total", netdata.IncrementalAlgorithm)
		authsChart.AddDimension("auth_errors", "errors", netdata.IncrementalAlgorithm)
		worker.AddChart(authsChart, collector)

		netChart := netdata.NewChart(instance, "net", "", "Network", "bytes", instance, "memcached.net")
		netChart.AddDimension("bytes_read", "in", netdata.IncrementalAlgorithm)
		netChart.AddDimension("bytes_written", "out", netdata.IncrementalAlgorithm)
		worker.AddChart(netChart, collector)

		lruChart := netdata.NewChart(instance, "lru", "", "LRU", "items", instance, "memcached.lru")
		lruChart.AddDimension("expired_unfetched", "expired_unfetched", netdata.IncrementalAlgorithm)
		lruChart.AddDimension("evicted_unfetched", "evicted_unfetched", netdata.IncrementalAlgorithm)
		lruChart.AddDimension("evicted_active", "evicted_active", netdata.IncrementalAlgorithm)
		lruChart.AddDimension("moves_to_cold", "moves_to_cold", netdata.IncrementalAlgorithm)
		lruChart.AddDimension("moves_to_warm", "moves_to_warm", netdata.IncrementalAlgorithm)
		lruChart.AddDimension("moves_within_lru", "moves_within_lru", netdata.IncrementalAlgorithm)
		worker.AddChart(lruChart, collector)
	}

	worker.Run()
}
