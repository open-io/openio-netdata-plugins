package openio

import(
    "bufio"
    "strings"
    "path"
    "os"
    "encoding/json"
    "fmt"
    "oionetdata/util"
    "oionetdata/netdata"
)

type serviceType []string

type serviceInfo []struct {
    Addr string
    Score int
    Local bool
}

var mReplacer = strings.NewReplacer(
    ".", "_",
    ":", "_",
    "/", "_",
)

/*
ProxyURL - Get URL of oioproxy from configuration
*/
func ProxyURL(basePath string, ns string) string {
  file, err := os.Open(path.Join(basePath, ns))
  util.RaiseIf(err)
  defer file.Close()

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
	t := scanner.Text()
	if strings.HasPrefix(t, "proxy") {
		return strings.Split(t, "=")[1];
	}
  }
  util.RaiseIf(scanner.Err())
  return ""
}

/*
Collect - collect openio metrics
*/
func Collect(proxyURL string, ns string, c chan netdata.Metric) {
	var sType = serviceTypes(proxyURL, ns)
	for t := range sType  {
		var sInfo = collectScore(proxyURL, ns, sType[t], c)
		if sType[t] == "rawx" {
			for sc := range sInfo {
				if (sInfo[sc].Local) {
					go collectRawx(ns, sInfo[sc].Addr, c)
				}
			}
		} else if strings.HasPrefix(sType[t], "meta") {
			for sc := range sInfo {
				if (sInfo[sc].Local) {
					go collectMetax(ns, sInfo[sc].Addr, proxyURL, c)
				}
			}
		}
	}
}

func serviceTypes(proxyURL string, ns string) serviceType {
	url := fmt.Sprintf("http://%s/v3.0/%s/conscience/info?what=types", proxyURL, ns)
	res := serviceType{}
	util.RaiseIf(json.Unmarshal([]byte(util.HTTPGet(url)), &res))
	return res
}

/*
CollectRawx - update metrics for Rawx services
*/
func collectRawx(ns string, service string, c chan netdata.Metric) {
	url := fmt.Sprintf("http://%s/stat", service)
	var lines = strings.Split(util.HTTPGet(url), "\n");
	for i := range lines {
		s := strings.Split(lines[i], " ")
		if s[0] == "counter" {
			netdata.Update(s[1], sID(service, ns), s[2], c)
		} else if s[1] == "volume" {
			go volumeInfo(service, ns, s[2], c)
		}
	}
}

/*
CollectMetax - update metrics for M0/M1/M2 servicess
*/
func collectMetax(ns string, service string, proxyURL string, c chan netdata.Metric) {
	url := fmt.Sprintf("http://%s/v3.0/forward/stats?id=%s", proxyURL, service)
	var lines = strings.Split(util.HTTPGet(url), "\n");
	for i := range lines {
		s := strings.Split(lines[i], " ")
		if s[0] == "counter" {
			netdata.Update(s[1], sID(service, ns), s[2], c)
		} else if s[1] == "volume" {
            go volumeInfo(service, ns, s[2], c)
		} else if s[0] == "gauge" {
			// TODO: do something with gauge?
		}
	}
}

func volumeInfo(service string, ns string, volume string, c chan netdata.Metric) {
    for dim, val := range util.VolumeInfo(volume) {
        netdata.Update(dim, sID(service, ns, volume), fmt.Sprint(val), c)
    }
}

/*
CollectScore - collect score values on all scored services
*/
func collectScore(proxyURL string, ns string, sType string, c chan netdata.Metric) (serviceInfo) {
	sInfo := serviceInfo{}
	url := fmt.Sprintf("http://%s/v3.0/%s/conscience/list?type=%s", proxyURL, ns, sType)
	util.RaiseIf(json.Unmarshal([]byte(util.HTTPGet(url)), &sInfo))
	for i := range sInfo {
        if util.IsSameHost(sInfo[i].Addr) {
            sInfo[i].Local = true
            netdata.Update("score", sID(sInfo[i].Addr, ns), fmt.Sprint(sInfo[i].Score), c)
        } else {
            sInfo[i].Local = false
        }
	}
	return sInfo
}

func sID(service string, ns string, volume ...string) (string) {
    if (len(volume) > 0) && (volume[0] != "") {
        return fmt.Sprintf("%s.%s.%s" , ns, mReplacer.Replace(service), volume[0]);
    }
    return fmt.Sprintf("%s.%s", ns, mReplacer.Replace(service));
}
