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

package container

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/open-io/openio-netdata-plugins/netdata"
	"github.com/open-io/openio-netdata-plugins/util"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var scriptListCont = redis.NewScript(`
    local res = {}
    local acct_key = "containers:" .. KEYS[1]
    local cont_pfix = "container:" .. KEYS[1] .. ":"
    local cont = redis.call("ZRANGE", acct_key, ARGV[2], ARGV[3])
    for _, c in ipairs(cont) do
        local k = cont_pfix .. c;
        local v = redis.call('HGET', k, 'objects')
        if v then
            if tonumber(v) > tonumber(ARGV[1]) then
                res[c] = {
									objects = string.format("%d", v),
									kbytes  = string.format("%d", redis.call('HGET', k, 'bytes')),
								}
            end;
        end;
    end;
    return cjson.encode(res);
`)

var scriptBucketInfo = redis.NewScript(`
	local buckets = redis.call("keys", "bucket:*");
	local res = {}
	for i, bucket in ipairs(buckets) do
      res[bucket] = {
          account = redis.call("hget", bucket, "account"),
          objects = string.format("%d", redis.call("hget", bucket, "objects")),
          kbytes  = string.format("%d", redis.call("hget", bucket, "bytes") / 1000),
      }
	end;
	return cjson.encode(res);
`)

var scriptAcctInfo = redis.NewScript(`
	local accts = redis.call("hgetall", "accounts:")
	local res = {}
	for i, acc in ipairs(accts) do
        if i % 2 == 1 then
					local acct_key = "account:" .. acc;
					local key = "containers:" .. acc;
					res[acc] = {
							objects = string.format("%d", redis.call("HGET", acct_key, 'objects')),
							kbytes  = string.format("%d", redis.call('HGET', acct_key, 'bytes') / 1000),
							count   = string.format("%d", redis.call("ZCOUNT", key, 0, 0)),
					}
        end;
	end;
	return cjson.encode(res);
`)

// RedisAddr -- get redis address
func RedisAddr(basePath string, ns string) (string, error) {
	p := path.Join(basePath, ns, "redis-*/redis.conf")
	match, err := filepath.Glob(p)
	if err != nil || len(match) == 0 {
		return "", fmt.Errorf("failed to find redis conf in %s", p)
	}

	conf, err := util.ReadConf(match[0], " ")
	if err != nil {
		return "", err
	}

	ip := conf["bind"]
	port := conf["port"]
	if len(ip) != 0 && len(port) != 0 {
		return fmt.Sprintf("%s:%s", ip, port), nil
	}
	return "", fmt.Errorf("invalid redis conf")
}

type infoObj struct {
	Account string `json:"account"`
	Objects string `json:"objects"`
	KBytes  string `json:"kbytes"`
	Count   string `json:"count"`
}

// Collect -- collect container metrics
func Collect(client *redis.Client, ns string, l int64, t int64, f bool, c chan netdata.Metric) error {
	bucketInfoStr, err := scriptBucketInfo.Run(client, []string{}, 0).Result()
	if err != nil {
		return err
	}
	bucketInfo := map[string]infoObj{}
	if err := json.Unmarshal([]byte(bucketInfoStr.(string)), &bucketInfo); err != nil {
		return err
	}
	for name, info := range bucketInfo {
		bucket := info.Account + "." + strings.Split(name, ":")[1]
		netdata.Update("account_bucket_kilobytes", bucket, info.KBytes, c)
		netdata.Update("account_bucket_objects", bucket, info.Objects, c)
	}

	acctInfo, err := scriptAcctInfo.Run(client, []string{}, 0).Result()
	if err != nil {
		return err
	}
	acctObj := map[string]infoObj{}
	if err := json.Unmarshal([]byte(acctInfo.(string)), &acctObj); err != nil {
		return err
	}
	for acc, info := range acctObj {
		// Note(VDO): the metric account_bytes can result in false results and has been deprecated
		netdata.Update("account_kilobytes", util.AcctID(ns, acc), info.KBytes, c)
		netdata.Update("account_objects", util.AcctID(ns, acc), info.Objects, c)
		netdata.Update("container_count", util.AcctID(ns, acc), info.Count, c)

		if count, err := strconv.ParseInt(info.Count, 10, 64); !f && err == nil {
			var i int64
			for i > int64(-1) && i < count {
				res, err := scriptListCont.Run(client, []string{acc}, t, i, l).Result()
				if err != nil {
					return err
				}
				contObj := map[string]infoObj{}
				if err := json.Unmarshal([]byte(res.(string)), &contObj); err != nil {
					return err
				}
				for cont, data := range contObj {
					netdata.Update("container_objects", util.AcctID(ns, acc, cont), data.Objects, c)
					netdata.Update("container_bytes", util.AcctID(ns, acc, cont), data.KBytes, c)
				}
				i += l
			}
		}
	}
	return nil
}
