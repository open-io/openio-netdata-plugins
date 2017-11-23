package main

import (
	"fmt"
	"net/http"
	"bufio"
	"os"
	"strings"
	"log"
	"io/ioutil"
	"encoding/json"
	"time"
	"flag"
	"strconv"
	"path"
)

// BasePath -- base configuration path
var BasePath = "/etc/oio/sds.conf.d/";

// PollsBeforeReload -- Number of cycles before collector is reloaded
// Restarting the collector every now and then should help getting rid of memleaks
var PollsBeforeReload = 1000;

// PollInterval -- Seconds to wait between cycles
var PollInterval int64 = 10;

// DimPrefix -- prefix to add to dimensions
var DimPrefix = "openio"

// ServiceType -- list of service types
type ServiceType []string

// Charts -- list of already created charts with the dimensions
var Charts = make(map[string]map[string]bool)

// Buffer -- metric buffer to be sent
var Buffer = make(map[string][]string)

// ServiceInfo -- decoded service metric information
type ServiceInfo []struct {
	Addr  string
	Score int
	Tags struct {
		Up bool
		Loc string
		Vol string
	}
}

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
		proxyURLs[namespaces[i]] = getProxyURL(namespaces[i]);
	}

	poll := 0

	for poll < PollsBeforeReload {
		for ns, proxyURL := range proxyURLs {
			updateServices(proxyURL, ns);
		}
		sendBuffer()
		time.Sleep(time.Duration(PollInterval) * 1000 * time.Millisecond);
		poll++;
	}
}

func raiseIf(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func httpGet(url string) string {
	resp, err := http.Get(url);
	raiseIf(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	raiseIf(err)
	return string(body)
}

func getProxyURL(ns string) string {
  file, err := os.Open(path.Join(BasePath, ns))
  raiseIf(err)
  defer file.Close()

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
	t := scanner.Text()
	if strings.HasPrefix(t, "proxy") {
		return strings.Split(t, "=")[1];
	}
  }
  raiseIf(scanner.Err())
  return ""
}

func createChart(chart string, desc string, title string, units string, family string) {
	fmt.Printf("CHART %s '%s' '%s' '%s' '%s'\n", chart, desc, title, units, family)
}

func createDim(dim string) {
	fmt.Printf("DIMENSION %s '%s' absolute\n", dim, dim)
}

func updateChart(chart string, dim string, value string) {
	dim = strings.Replace(dim, ".", "_", -1)
	dim = strings.Replace(dim, ":", "_", -1)
	chart = fmt.Sprintf("%s.%s", DimPrefix, strings.Replace(chart, ".", "_", -1))
	chartTitle := strings.ToUpper(strings.Join(strings.Split(chart, "_"), " "))
	if _, e := Charts[chart]; !e {
		createChart(chart, "", chartTitle, "", strings.Split(chart, ".")[1])
		Charts[chart] = make(map[string]bool)
	}
	if _, e := Charts[chart][dim]; !e {
		createChart(chart, "", chartTitle, "", strings.Split(chart, ".")[1])
		createDim(dim)
		Charts[chart][dim] = true
	}

	Buffer[chart]=append(Buffer[chart], fmt.Sprintf("SET %s %s", dim, value))
}

func sendBuffer() {
	for chart := range Buffer {
		fmt.Printf("BEGIN %s\n", chart)
		for _, v := range Buffer[chart] {
			fmt.Println(v)
		}
 		fmt.Println("END")
	}
	// Send & reset the buffer after the collection
	Buffer = make(map[string][]string)
}

func getServiceTypes(proxyURL string, ns string) ServiceType {
	url := fmt.Sprintf("http://%s/v3.0/%s/conscience/info?what=types", proxyURL, ns)
	res := ServiceType{}
	raiseIf(json.Unmarshal([]byte(httpGet(url)), &res))
	return res
}

func updateRawxCounters(ns string, service string) {
	url := fmt.Sprintf("http://%s/stat", service)
	var lines = strings.Split(httpGet(url), "\n");
	for i := range lines {
		s := strings.Split(lines[i], " ")
		if s[0] == "counter" {
			updateChart(s[1], fmt.Sprintf("%s.%s", service, ns), s[2])
		} else if s[1] == "volume" {
			// TODO: do something with volume?
		}
	}
}

func updateMetaxCounters(ns string, service string, proxyURL string) {
	url := fmt.Sprintf("http://%s/v3.0/forward/stats?id=%s", proxyURL, service)
	var lines = strings.Split(httpGet(url), "\n");
	for i := range lines {
		s := strings.Split(lines[i], " ")
		if s[0] == "counter" {
			updateChart(s[1], fmt.Sprintf("%s.%s", service, ns), s[2])
		} else if s[1] == "volume" {
			// TODO: do something with volume?
		} else if s[1] == "gauge" {
			// TODO: do something with gauge?
		}
	}
}

func updateScore(proxyURL string, ns string, serviceType string) ServiceInfo {
	serviceInfo := ServiceInfo{}
	url := fmt.Sprintf("http://%s/v3.0/%s/conscience/list?type=%s", proxyURL, ns, serviceType)
	raiseIf(json.Unmarshal([]byte(httpGet(url)), &serviceInfo))
	for i := range serviceInfo {
		updateChart("score", fmt.Sprintf("%s.%s", serviceInfo[i].Addr, ns), fmt.Sprint(serviceInfo[i].Score))
	}
	return serviceInfo
}

func updateServices(proxyURL string, ns string) {
	var serviceType = getServiceTypes(proxyURL, ns)
	for t := range serviceType  {
		var serviceInfo = updateScore(proxyURL, ns, serviceType[t])
		if serviceType[t] == "rawx" {
			for sc := range serviceInfo {
				if strings.HasPrefix(serviceInfo[sc].Addr, strings.Split(proxyURL, ":")[0]) {
					updateRawxCounters(ns, serviceInfo[sc].Addr)
				}
			}
		} else if strings.HasPrefix(serviceType[t], "meta") {
			for sc := range serviceInfo {
				if strings.HasPrefix(serviceInfo[sc].Addr, strings.Split(proxyURL, ":")[0]) {
					updateMetaxCounters(ns, serviceInfo[sc].Addr, proxyURL)
				}
			}
		}
	}
}
