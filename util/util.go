package util

import(
    "syscall"
    "log"
    "net/http"
    "io/ioutil"
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

/*
RaiseIf - Exit with error
*/
func RaiseIf(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
