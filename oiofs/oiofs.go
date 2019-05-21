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

package oiofs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// Endpoint defined an oiofs endpoint to monitor
type Endpoint struct {
	URL  string
	Path string
}

// Ops defines all possible metrics
var Ops = map[string][]string{
	"metaPrefix": []string{"Meta"},
	"metaDebug": []string{
		"addDir", "addLink", "allocateInode", "checkFsExists", "deallocateInode", "delDir", "delLink", "deleteFs",
		"getInodeStat", "getXAttr", "incrNlink", "listXAttr", "lookupInodeStat", "maxIno", "mkfs", "readdir",
		"removeXAttr", "setInodeStat", "setLink", "setSymlink", "setXAttr", "updateTimestampsInode", "init_ctx",
	},
	"fusePrefix": []string{"fuse"},
	"fuse":       []string{"read_total_byte", "write_total_byte"},
	"fuseDebug": []string{
		" rename", "create", "fallocate", "flush", "forget", "fsync", "getattr", "getxattr",
		"link", "listxattr", "lookup", "mkdir", "mknod", "open", "opendir", "read", "readdir", "readlink", "release",
		"releasedir", "rmdir", "setattr", "setxattr", "statfs", "symlink", "unlink", "write",
	},
	"sdsPrefix": []string{"sds"},
	"sds": []string{"upload_failed", "upload_succeeded", "download_failed", "download_succeeded",
		"download_total_byte", "upload_total_byte",
	},
	"sdsDebug": []string{
		"delete", "deleteAllContainers", "deleteFs", "flushContainer", "mkfs", "pad", "replace",
		"replaceChunk", "replacePartialChunk", "statFs", "truncate", "download", "upload",
	},
	"cachePrefix": []string{"cache"},
	"cache": []string{
		"chunk_count", "read_count", "read_total_us", "read_count", "read_hit",
		"read_miss", "read_total_us", "chunk_avg_age_microseconds", "read_total_byte",
		"chunk_total_byte", "chunk_used_byte",
	},
	"cacheDebug": []string{},
}

type collector struct {
	endpoint  Endpoint
	full      bool
	whitelist map[string]int64
}

func NewCollector(endpoint Endpoint, full bool) *collector {
	var whitelist = map[string]int64{}
	// Generate whitelist of metrics to keep
	for _, p := range []string{"meta", "sds", "fuse", "cache"} {
		var prefix = Ops[fmt.Sprintf("%sPrefix", p)][0]
		for _, op := range Ops[p] {
			whitelist[fmt.Sprintf("%s_%s", prefix, op)] = 0
		}
		if full {
			for _, op := range Ops[fmt.Sprintf("%sDebug", p)] {
				if p == "sds" || (p == "meta" && op != "init_ctx") {
					op = strings.Title(op)
				}
				whitelist[fmt.Sprintf("%s_%s_count", prefix, op)] = 0
				whitelist[fmt.Sprintf("%s_%s_total_us", prefix, op)] = 0
			}
		}
	}

	return &collector{
		endpoint:  endpoint,
		full:      full,
		whitelist: whitelist,
	}
}

func (c *collector) Collect() (map[string]string, error) {
	// TODO: support v2

	r, err := http.Get(fmt.Sprintf("http://%s/stats", c.endpoint.URL))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var rs interface{}
	if err := json.Unmarshal(body, &rs); err != nil {
		return nil, err
	}

	var res = map[string]string{}

	for k := range c.whitelist {
		// NOTE: VDO: the cost of type assertion is negligible
		if v, ok := rs.(map[string]interface{})[k]; ok && v != nil {
			res[k] = strconv.FormatInt(int64(v.(float64)), 10)
		} else {
			res[k] = "0"
		}
	}

	return res, nil
}
