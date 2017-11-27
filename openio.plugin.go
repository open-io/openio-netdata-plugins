package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"flag"
	"strconv"
	"oionetdata/openio"
	"oionetdata/netdata"
)

// BasePath -- base configuration path
var BasePath = "/etc/oio/sds.conf.d/";

// PollsBeforeReload -- Number of cycles before collector is reloaded
// Restarting the collector every now and then should help getting rid of memleaks
var PollsBeforeReload = 1000;

// PollInterval -- Seconds to wait between cycles
var PollInterval int64 = 10;

func main() {
	if len(os.Args) > 1 {
		var err error
		PollInterval, err = strconv.ParseInt(os.Args[1], 10, 0)
		if err == nil {
			os.Args = append(os.Args[:1], os.Args[2:]...)
		}
	}

	nsPtr := flag.String("ns", "OPENIO", "List of namespaces delimited by semicolons (:)")
	confPtr := flag.String("conf", "/etc/oio/sds.conf.d/", "Path to SDS config")
	flag.Parse()

	BasePath = *confPtr

	var proxyURLs = make(map[string]string)
	var namespaces = strings.Split(*nsPtr, ":")
	for i := range namespaces {
		proxyURLs[namespaces[i]] = openio.ProxyURL(BasePath, namespaces[i]);
	}

	poll := 0

	for poll < PollsBeforeReload {
		for ns, proxyURL := range proxyURLs {
			openio.Collect(proxyURL, ns);
		}
		sendBuffer()
		time.Sleep(time.Duration(PollInterval) * 1000 * time.Millisecond);
		poll++;
	}
}

func sendBuffer() {
	for chart := range netdata.Buffer {
		fmt.Printf("BEGIN %s\n", chart)
		for _, v := range netdata.Buffer[chart] {
			fmt.Println(v)
		}
 		fmt.Println("END")
	}
	// Send & reset the buffer after the collection
	netdata.Buffer = make(map[string][]string)
}
