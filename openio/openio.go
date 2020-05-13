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

package openio

import (
	"encoding/json"
	"fmt"
	"log"
	"oionetdata/netdata"
	"oionetdata/util"
	"path"
	"strconv"
	"strings"
	"sync"
)

type serviceType []string

type serviceInfo []struct {
	Addr  string
	Score int
	Local bool
}

type cntr struct {
	sync.RWMutex
	counter map[string]int
}

func makeCounter() *cntr {
	return &cntr{
		counter: make(map[string]int),
	}
}

func (c *cntr) setCounter(key string, value int) {
	c.Lock()
	defer c.Unlock()
	c.counter[key] = value
}

func (c *cntr) getCounter(key string) (int, bool) {
	c.RLock()
	defer c.RUnlock()
	v, ok := c.counter[key]
	return v, ok
}

var counter = makeCounter()

// CollectInterval -- collection interval for derivatives
var CollectInterval = 10

// ProxyAddr returns the proxy address from namespace configuration
func ProxyAddr(basePath string, ns string) (string, error) {
	conf, err := util.ReadConf(path.Join(basePath, ns), "=")
	if err != nil {
		return "", err
	}
	addr := conf["proxy"]
	if len(addr) != 0 {
		return addr, nil
	}
	return "", fmt.Errorf("no proxy address found for %s", ns)
}

// ZookeeperAddr retrieves local zookeeper address from namespace configuration
func ZookeeperAddr(basePath string, ns string) (string, error) {
	const clusterSep = ";"
	const addrSep = ","
	conf, err := util.ReadConf(path.Join(basePath, ns), "=")
	if err != nil {
		return "", err
	}
	zkStr := conf["zookeeper"]
	if len(zkStr) != 0 {
		for _, zkCluster := range strings.Split(conf["zookeeper"], clusterSep) {
			for _, zkAddr := range strings.Split(zkCluster, addrSep) {
				if util.IsSameHost(zkAddr) {
					return zkAddr, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no local zookeeper address found for %s", ns)
}

func diffCounter(metric string, sid string, value string) string {
	res := ""
	curr, err := strconv.Atoi(value)
	if err != nil {
		return res
	}

	if prev, ok := counter.getCounter(metric + sid); ok {
		res = strconv.Itoa((curr - prev) / CollectInterval)
	}
	counter.setCounter(metric+sid, curr)
	return res
}

/*
Collect - collect openio metrics
*/
func Collect(proxyURL string, ns string, c chan netdata.Metric) {
	sType, err := serviceTypes(proxyURL, ns)
	if err != nil {
		log.Println("WARN: couldn't retrieve service types", err)
		return
	}
	for t := range sType {
		sInfo, err := collectScore(proxyURL, ns, sType[t], c)
		if err != nil {
			log.Println("WARN: could not retrieve services of type", sType[t], err)
			continue
		}
		if sType[t] == "rawx" {
			for sc := range sInfo {
				if sInfo[sc].Local {
					go collectRawx(ns, sInfo[sc].Addr, c)
				}
			}
		} else if strings.HasPrefix(sType[t], "meta") {
			for sc := range sInfo {
				if sInfo[sc].Local {
					go collectMetax(ns, sInfo[sc].Addr, proxyURL, c)
					if sType[t] == "meta2" {
						go collectMeta2Info(ns, sInfo[sc].Addr, proxyURL, c)
					}
				}
			}
		}
	}
}

func serviceTypes(proxyURL string, ns string) (serviceType, error) {
	url := fmt.Sprintf("http://%s/v3.0/%s/conscience/info?what=types", proxyURL, ns)
	res := serviceType{}

	typesResponse, err := util.HTTPGet(url)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(typesResponse), &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func collectRawx(ns string, service string, c chan netdata.Metric) {
	url := fmt.Sprintf("http://%s/stat", service)
	res, err := util.HTTPGet(url)
	if err != nil {
		log.Println("WARN: rawx metric collection failed", err)
		return
	}
	var lines = strings.Split(res, "\n")
	for i := range lines {
		s := strings.Split(lines[i], " ")
		if len(s) != 3 {
			continue
		}
		if s[0] == "counter" {
			if diff := diffCounter(s[1], util.SID(service, ns), s[2]); diff != "" {
				netdata.Update(s[1], util.SID(service, ns), diff, c)
			}
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
	res, err := util.HTTPGet(url)
	if err != nil {
		log.Println("WARN: metaX stats collection failed", err)
		return
	}
	var lines = strings.Split(res, "\n")
	for i := range lines {
		s := strings.Split(lines[i], " ")
		if s[0] == "counter" {
			if diff := diffCounter(s[1], util.SID(service, ns), s[2]); diff != "" {
				netdata.Update(s[1], util.SID(service, ns), diff, c)
			}
		} else if s[1] == "volume" {
			go volumeInfo(service, ns, s[2], c)
		}
	}
}

type metaxInfoBody struct {
	Cache     map[string]int64
	Elections map[string]int64
}

func collectMeta2Info(ns, service, proxyURL string, c chan netdata.Metric) {
	url := fmt.Sprintf("http://%s/v3.0/forward/info?id=%s", proxyURL, service)
	sid := util.SID(service, ns)
	info := metaxInfoBody{}

	res, err := util.HTTPGet(url)

	if err != nil {
		log.Println("WARN: meta2 info collection failed", err)
		return
	}

	if err = json.Unmarshal([]byte(res), &info); err != nil {
		log.Println("WARN: meta2 info collection failed", err)
		return
	}

	for dim, val := range info.Cache {
		netdata.Update(fmt.Sprintf("meta2_cache_bases_%s", dim), sid, fmt.Sprint(val), c)
	}
	for dim, val := range info.Elections {
		netdata.Update(fmt.Sprintf("meta2_elections_%s", dim), sid, fmt.Sprint(val), c)
	}
}

func volumeInfo(service string, ns string, volume string, c chan netdata.Metric) {
	info, fsid, err := util.VolumeInfo(volume)
	if err != nil {
		log.Println("WARN: volume info collection failed", err)
		return
	}
	for dim, val := range info {
		netdata.Update(dim, util.SID(service, ns, fsid), fmt.Sprint(val), c)
	}
}

func collectScore(proxyURL string, ns string, sType string, c chan netdata.Metric) (serviceInfo, error) {
	sInfo := serviceInfo{}
	url := fmt.Sprintf("http://%s/v3.0/%s/conscience/list?type=%s", proxyURL, ns, sType)
	res, err := util.HTTPGet(url)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(res), &sInfo)
	if err != nil {
		return nil, err
	}
	for i := range sInfo {
		if util.IsSameHost(sInfo[i].Addr) {
			sInfo[i].Local = true
			netdata.Update("score", util.SID(sType+"_"+sInfo[i].Addr, ns), fmt.Sprint(sInfo[i].Score), c)
		} else {
			sInfo[i].Local = false
		}
	}
	return sInfo, nil
}
