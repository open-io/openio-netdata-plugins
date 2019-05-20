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

package util

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var ipList map[string]bool

// ForceRemote -- force remote metric collection
var ForceRemote = false

var mReplacer = strings.NewReplacer(
	".", "_",
	":", "_",
	"/", "_",
)

// VolumeInfo retrieves metrics from statfs
func VolumeInfo(volume string) (map[string]uint64, string, error) {
	f, err := os.OpenFile(volume, os.O_RDONLY, 0644)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	var stat syscall.Statfs_t
	err = syscall.Fstatfs(int(f.Fd()), &stat)
	if err != nil {
		return nil, "", err
	}
	vMetric := make(map[string]uint64)
	vMetric["byte_avail"] = stat.Bavail * uint64(stat.Bsize)
	vMetric["byte_used"] = (stat.Blocks - stat.Bfree) * uint64(stat.Bsize)
	vMetric["byte_free"] = stat.Bfree * uint64(stat.Bsize)
	vMetric["inodes_free"] = stat.Ffree
	vMetric["inodes_used"] = stat.Files - stat.Ffree

	fsId := strconv.FormatUint(uint64(stat.Fsid.X__val[1])<<32|uint64(stat.Fsid.X__val[0]), 10)
	return vMetric, fsId, nil
}

/*
HTTPGet - Wrapper for Get HTTP request
*/
func HTTPGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getIPList() map[string]bool {
	ipList := make(map[string]bool)
	ifaces, err := net.InterfaceAddrs()
	RaiseIf(err)
	for ip := range ifaces {
		// TODO: consider adding support for IPv6 here
		ips := strings.Split(ifaces[ip].String(), "/")[0]
		if net.ParseIP(ips).To4() != nil {
			ipList[ips] = true
		}
	}
	return ipList
}

// IsSameHost -- checks if a service if on the current host
func IsSameHost(service string) bool {
	if ForceRemote {
		return true
	}
	if ipList == nil {
		ipList = getIPList()
	}
	serviceIP := strings.Split(service, ":")[0]
	_, ok := ipList[serviceIP]
	return ok
}

/*
RaiseIf - Exit with error
*/
func RaiseIf(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// SID -- Get a service ID for netdata
func SID(service string, ns string, volume ...string) string {
	if (len(volume) > 0) && (volume[0] != "") {
		return fmt.Sprintf("%s.%s.%s", ns, mReplacer.Replace(service), volume[0])
	}
	return fmt.Sprintf("%s.%s", ns, mReplacer.Replace(service))
}

// AcctID -- get an ID from a ns/account/container
func AcctID(ns string, acct string, cont ...string) string {
	if (len(cont) > 0) && (cont[0] != "") {
		return fmt.Sprintf("%s.%s.%s", ns, mReplacer.Replace(acct), mReplacer.Replace(cont[0]))
	}
	return fmt.Sprintf("%s.%s", ns, mReplacer.Replace(acct))
}

// Commands retrieves commands to execute on each node
func Commands(path string) (map[string]string, error) {
	conf, err := ReadConf(path, "=")
	if err != nil {
		return nil, err
	}
	return conf, err
}

// OiofsEndpoints retrieves oiofs endpoints to monitor
func OiofsEndpoints(path string) (map[string]string, error) {
	conf, err := ReadConf(path, "=")
	if err != nil {
		return nil, err
	}
	return conf, err
}

// S3RoundtripConfig returns the S3 credentials and endpoint
func S3RoundtripConfig(path string) (map[string]string, error) {
	conf, err := ReadConf(path, "=")
	if err != nil {
		return nil, err
	}
	return conf, err
}

func ReadConf(path string, separator string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if pos := strings.Index(line, separator); pos > 0 {
			if key := strings.TrimSpace(line[:pos]); len(key) > 0 {
				value := ""
				if len(line) > pos {
					value = strings.TrimSpace(line[pos+1:])
				}
				config[key] = value
			}
		}
	}
	return config, nil
}
