package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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

	zkType := fmt.Sprintf("zk_%s_%s", ns, addr)
	family := "zookeeper"

	// Latency
	latencyStats := netdata.NewChart(zkType, "latency", "", "Latency Stats", "microseconds", family)
	latencyStats.AddDimension("zk_min_latency", "min", netdata.AbsoluteAlgorithm)
	latencyStats.AddDimension("zk_max_latency", "max", netdata.AbsoluteAlgorithm)
	latencyStats.AddDimension("zk_avg_latency", "avg", netdata.AbsoluteAlgorithm)
	worker.AddChart(latencyStats)

	// Packets
	packetsStats := netdata.NewChart(zkType, "packets", "", "Packets Stats", "packets/s", family)
	packetsStats.AddDimension("zk_packets_received", "received", netdata.IncrementalAlgorithm)
	packetsStats.AddDimension("zk_packets_sent", "sent", netdata.IncrementalAlgorithm)
	worker.AddChart(packetsStats)

	// Connections
	connectionsStats := netdata.NewChart(zkType, "connections", "", "Connections Stats", "connections", family)
	connectionsStats.AddDimension("zk_num_alive_connections", "alive", netdata.AbsoluteAlgorithm)
	worker.AddChart(connectionsStats)

	// Requests
	requestsStats := netdata.NewChart(zkType, "requests", "", "Requests Stats", "requests", family)
	requestsStats.AddDimension("zk_outstanding_requests", "outstanding", netdata.AbsoluteAlgorithm)
	worker.AddChart(requestsStats)

	// Nodes
	nodesStats := netdata.NewChart(zkType, "nodes", "", "Nodes Stats", "nodes", family)
	nodesStats.AddDimension("zk_znode_count", "znode", netdata.AbsoluteAlgorithm)
	nodesStats.AddDimension("zk_watch_count", "watch", netdata.AbsoluteAlgorithm)
	nodesStats.AddDimension("zk_ephemerals_count", "ephemeral", netdata.AbsoluteAlgorithm)
	worker.AddChart(nodesStats)

	// Data
	dataStats := netdata.NewChart(zkType, "data", "", "Data Stats", "bytes", family)
	dataStats.AddDimension("zk_approximate_data_size", "size", netdata.AbsoluteAlgorithm)
	worker.AddChart(dataStats)

	// File descriptors
	fdStats := netdata.NewChart(zkType, "fds", "", "File descriptors", "fds", family)
	fdStats.AddDimension("zk_open_file_descriptor_count", "open", netdata.AbsoluteAlgorithm)
	fdStats.AddDimension("zk_max_file_descriptor_count", "max", netdata.AbsoluteAlgorithm)
	worker.AddChart(fdStats)

	worker.Run()
}
