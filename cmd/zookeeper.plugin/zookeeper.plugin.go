package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"oionetdata/collector"
	"oionetdata/netdata"
	"oionetdata/openio"
	"oionetdata/zookeeper"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}
	var ns string
	var conf string
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&ns, "ns", "OPENIO", "Namespace")
	fs.StringVar(&conf, "conf", "/etc/oio/sds.conf.d/", "Path to SDS config")
	fs.Parse(os.Args[2:])
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	addr, err := openio.ZookeeperAddr(conf, ns)
	if err != nil {
		log.Fatalf("Load failure: %v", err)
	}
	writer := netdata.NewDefaultWriter()
	collector := zookeeper.NewCollector(addr)
	worker := netdata.NewWorker(time.Duration(intervalSeconds)*time.Second, writer, collector)

	fAddr := strings.Replace(addr, ".", "_", -1)
	fAddr = strings.Replace(addr, ":", "_", -1)
	zkType := fmt.Sprintf("zk_%s_%s", ns, fAddr)
	family := "zookeeper"

	// Latency
	latencyStats := netdata.NewChart(zkType, "latency", "", "Latency Stats", "microseconds", family, "zk.latency")
	latencyStats.AddDimension("zk_min_latency", "min", netdata.AbsoluteAlgorithm)
	latencyStats.AddDimension("zk_max_latency", "max", netdata.AbsoluteAlgorithm)
	latencyStats.AddDimension("zk_avg_latency", "avg", netdata.AbsoluteAlgorithm)
	worker.AddChart(latencyStats)

	// Packets
	packetsStats := netdata.NewChart(zkType, "packets", "", "Packets Stats", "packets/s", family, "zk.packets")
	packetsStats.AddDimension("zk_packets_received", "received", netdata.IncrementalAlgorithm)
	packetsStats.AddDimension("zk_packets_sent", "sent", netdata.IncrementalAlgorithm)
	worker.AddChart(packetsStats)

	// Connections
	connectionsStats := netdata.NewChart(zkType, "connections", "", "Connections Stats", "connections", family, "zk.connections")
	connectionsStats.AddDimension("zk_num_alive_connections", "alive", netdata.AbsoluteAlgorithm)
	worker.AddChart(connectionsStats)

	// Requests
	requestsStats := netdata.NewChart(zkType, "requests", "", "Requests Stats", "requests", family, "zk.requests")
	requestsStats.AddDimension("zk_outstanding_requests", "outstanding", netdata.AbsoluteAlgorithm)
	worker.AddChart(requestsStats)

	// Nodes
	nodesStats := netdata.NewChart(zkType, "nodes", "", "Nodes Stats", "nodes", family, "zk.nodes")
	nodesStats.AddDimension("zk_znode_count", "znode", netdata.AbsoluteAlgorithm)
	nodesStats.AddDimension("zk_watch_count", "watch", netdata.AbsoluteAlgorithm)
	nodesStats.AddDimension("zk_ephemerals_count", "ephemeral", netdata.AbsoluteAlgorithm)
	worker.AddChart(nodesStats)

	// Data
	dataStats := netdata.NewChart(zkType, "data", "", "Data Stats", "bytes", family, "zk.data")
	dataStats.AddDimension("zk_approximate_data_size", "size", netdata.AbsoluteAlgorithm)
	worker.AddChart(dataStats)

	// File descriptors
	fdStats := netdata.NewChart(zkType, "fds", "", "File descriptors", "fds", family, "zk.fds")
	fdStats.AddDimension("zk_open_file_descriptor_count", "open", netdata.AbsoluteAlgorithm)
	fdStats.AddDimension("zk_max_file_descriptor_count", "max", netdata.AbsoluteAlgorithm)
	worker.AddChart(fdStats)

	// (Leader) Pending syncs
	syncStats := netdata.NewChart(zkType, "syncs", "", "Pending syncs", "syncs", family, "zk.syncs")
	syncStats.AddDimension("zk_pending_syncs", "syncs", netdata.AbsoluteAlgorithm)
	worker.AddChart(syncStats)

	worker.Run()
}
