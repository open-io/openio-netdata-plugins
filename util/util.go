package util

import(
    "syscall"
    "log"
    "net/http"
    "io/ioutil"
    "net"
    "strings"
    "fmt"
)

var ipList map[string]bool

var mReplacer = strings.NewReplacer(
    ".", "_",
    ":", "_",
    "/", "_",
)

/*
VolumeInfo - Get volume metrics from statfs
*/
func VolumeInfo(volume string) (map[string]uint64) {
    var stat syscall.Statfs_t
    syscall.Statfs(volume, &stat)
    vMetric := make(map[string]uint64)
    vMetric["byte_avail"] = stat.Bavail * uint64(stat.Bsize)
    vMetric["byte_used"] = (stat.Blocks - stat.Bfree) * uint64(stat.Bsize)
    vMetric["byte_free"] = stat.Bfree * uint64(stat.Bsize)
    vMetric["inodes_free"] = stat.Ffree
    vMetric["inodes_used"] = stat.Files - stat.Ffree
    return vMetric
}

/*
HTTPGet - Wrapper for Get HTTP request
*/
func HTTPGet(url string) string {
	resp, err := http.Get(url);
	RaiseIf(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	RaiseIf(err)
	return string(body)
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
func IsSameHost(service string) (bool) {
    if ipList == nil {
        ipList = getIPList()
    }
    serviceIP := strings.Split(service, ":")[0]
    _, ok := ipList[serviceIP];
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
func SID(service string, ns string, volume ...string) (string) {
    if (len(volume) > 0) && (volume[0] != "") {
        return fmt.Sprintf("%s.%s.%s" , ns, mReplacer.Replace(service), volume[0]);
    }
    return fmt.Sprintf("%s.%s", ns, mReplacer.Replace(service));
}
