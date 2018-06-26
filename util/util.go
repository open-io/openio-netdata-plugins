package util

import (
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

/*
VolumeInfo - Get volume metrics from statfs
*/
func VolumeInfo(volume string) (map[string]uint64, string) {
	var stat syscall.Statfs_t
	f, err := os.OpenFile(volume, os.O_RDONLY, 0644)
	defer f.Close()
	RaiseIf(err)
	syscall.Fstatfs(int(f.Fd()), &stat)
	vMetric := make(map[string]uint64)
	vMetric["byte_avail"] = stat.Bavail * uint64(stat.Bsize)
	vMetric["byte_used"] = (stat.Blocks - stat.Bfree) * uint64(stat.Bsize)
	vMetric["byte_free"] = stat.Bfree * uint64(stat.Bsize)
	vMetric["inodes_free"] = stat.Ffree
	vMetric["inodes_used"] = stat.Files - stat.Ffree
	return vMetric, strconv.FormatUint(uint64(stat.Fsid.X__val[1])<<32|uint64(stat.Fsid.X__val[0]), 10)
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
