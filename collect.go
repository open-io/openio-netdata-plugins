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
	"strconv"
	//"time"
)

// BasePath -- base configuration path
var BasePath = "/tmp/sds.conf.d/";

// PollsBeforeReload -- Number of cycles before collector is reloaded
var PollsBeforeReload = 1000;

// PollInterval -- Seconds to wait between cycles
var PollInterval = 10;

func main() {
	ns := "OPENIO";
	proxyURL, err := getProxyURL(ns)
	raiseIf(err)

	// lastOp := time.Now()
	// elapsed := 0
	// now := time.Now()
	//createCharts(reqServices("rawx", proxyURL, ns))
	sTypes := reqInfo(proxyURL, ns)

	serviceList := reqServices(sTypes, proxyURL, ns)
	fmt.Printf("%+v\n", serviceList);
	return

	//poll := 0;

	// for poll < PollsBeforeReload {
	// 	serviceList := reqServices(sTypes, proxyURL, ns)
	// 	now = time.Now()
	// 	elapsed = int(now.Sub(lastOp) / 1000)
	// 	lastOp = now
	//
	// 	// Update charts
	// 	// fmt.Println(res2, elapsed, sTypes)
	//
	// 	for i := range serviceList {
	// 		updateChart(elapsed, serviceList[i].Addr, serviceList[i].Score)
    // 	}
	//
	// 	time.Sleep(time.Duration(PollInterval) * 1000 * time.Millisecond)
	// 	poll++;
	// }

}

// Initialize charts
func createCharts(serviceList ServiceInfo) {
	fmt.Println("CHART openio.score '' 'OpenIO Service score' 'score'")
	for i := range serviceList {
		fmt.Printf("DIMENSION %s '' absolute\n", serviceList[i].Addr)
	}
	// fmt.Println("DIMENSION service '' absolute")
	// fmt.Println("DIMENSION score '' absolute")
	// for _, v := range sTypes {
	// 	fmt.Printf("DIMENSION %s '' absolute\n", v)
	// }
}

func updateChart(elapsed int, service string, value int) {
	fmt.Printf("BEGIN openio.score %d\n", elapsed)
	fmt.Printf("SET %s %d\n", service, value)
	fmt.Println("END")
}

func makeRequest(url string) (string) {
	resp, err := http.Get(url);
	raiseIf(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	raiseIf(err)
	return string(body)
}

// ServiceType -- list of service types
type ServiceType []string

// ServiceInfo -- decoded service metric information
type ServiceInfo []struct {
	Addr  string
	Score int
	Tags struct {
		Up bool
		Loc string
		Vol string
	}
	Metax map[string]int64
	Rawx map[string]int64
	Volume string
}

// ServiceData -- assembled info of all service types
type ServiceData map[string]ServiceInfo




func reqInfo(proxyURL string, ns string) (ServiceType) {
	url := fmt.Sprintf("http://%s/v3.0/%s/conscience/info?what=types", proxyURL, ns)
	res := ServiceType{}
	json.Unmarshal([]byte(makeRequest(url)), &res)
	return res
}

func parseRawxCounters(body string) (map[string]int64, string) {
	res := make(map[string]int64)
	volume := ""
	var lines = strings.Split(body, "\n");
	for i := range lines {
		s := strings.Split(lines[i], " ")
		if s[0] == "counter" {
			v, err := strconv.ParseInt(s[2], 10, 64)
			raiseIf(err)
			res[s[1]] = v
		} else if s[1] == "volume" {
			volume = s[2]
		}
	}
	return res, volume
}

func reqServices(serviceType ServiceType, proxyURL string, ns string) (ServiceData) {
	res := ServiceData{}
	for i := range serviceType {
		serviceInfo := ServiceInfo{}
		fmt.Println(serviceType[i])
		url := fmt.Sprintf("http://%s/v3.0/%s/conscience/list?type=%s", proxyURL, ns, serviceType[i])
		json.Unmarshal([]byte(makeRequest(url)), &serviceInfo)

		if serviceType[i] == "rawx" {
			for j := range serviceInfo {
				url := fmt.Sprintf("http://%s/stat", serviceInfo[j].Addr)
				serviceInfo[j].Rawx, serviceInfo[j].Volume = parseRawxCounters(makeRequest(url))
				// serviceInfo[j].Rawx = rawxInfo
				// serviceInfo[j].Volume = volume
			}
		}

		res[serviceType[i]] = serviceInfo

	}
	return res
}

func raiseIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getProxyURL(ns string) (string, error) {
  file, err := os.Open(BasePath + ns)
  raiseIf(err)
  defer file.Close()

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
	t := scanner.Text()
	if strings.HasPrefix(t, "proxy") {
		return strings.Split(t, "=")[1], nil;
	}
  }
  return "", scanner.Err()
}
