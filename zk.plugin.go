package main

import (
	"strings"
	"strconv"
	// "oionetdata/openio"
	// "oionetdata/netdata"
    "fmt"
	"log"
    "net"
)

const req = "mntr\n"
var	ignore = map[string]bool{
	"zk_version":true,
	"zk_server_state":true,
}

func getStats(ip string, port int) {
	addr := strings.Join([]string{ip, strconv.Itoa(port)}, ":")
	conn, err := net.Dial("tcp", addr)

	defer conn.Close()

	if err != nil {
		log.Fatalln(err)
	}

	conn.Write([]byte(req))

	buff := make([]byte, 4096)
	n, _ := conn.Read(buff)
	parseStats(strings.Split(string(buff[:n-1]), "\n"))
}

func parseStats(stats []string) {
	for s := range stats {
			kv := strings.Split(stats[s], "\t")
			if _, o := ignore[kv[0]]; !o {
				fmt.Printf("Receive: %s %s\n", kv[0], kv[1])
			}
	}
}

func main() {

	var (
		ip   = "192.168.50.3"
		port = 6005
	)

	getStats(ip, port)

}


package main

import (
	"os"
	"strings"
	"flag"
	"oionetdata/openio"
	"oionetdata/netdata"
	"oionetdata/collector"
)

func main() {
	var interval int64;
	os.Args, interval = collector.ParseInterval(os.Args)
	nsPtr := flag.String("ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	confPtr := flag.String("conf", "/etc/oio/sds.conf.d/", "Path to SDS config")
	flag.Parse()

	var proxyURLs = make(map[string]string)
	var namespaces = strings.Split(*nsPtr, ":")
	for i := range namespaces {
		proxyURLs[namespaces[i]] = openio.ProxyURL(*confPtr, namespaces[i]);
	}

	collector.Run(interval, makeCollect(proxyURLs))
}

func makeCollect(proxyURLs map[string]string,) (collect collector.Collect) {
	return func(c chan netdata.Metric) {
		for ns, proxyURL := range proxyURLs {
			openio.Collect(proxyURL, ns, c);
		}
	}
}
