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
	"context"
	"encoding/json"
	"fmt"
	"oionetdata/netdata"
	"oionetdata/util"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

var scriptGetAccounts = redis.NewScript(`
    return redis.call("hgetall", "accounts:")
`)

var scriptGetContCount = redis.NewScript(`
    local key = "containers:" .. KEYS[1]
    return redis.call("ZCOUNT", key, 0, 0)
`)

var scriptListCont = redis.NewScript(`
    local res = {}
    local acct_key = "containers:" .. KEYS[1]
    local cont_pfix = "container:" .. KEYS[1] .. ":"
    local cont = redis.call("ZRANGE", acct_key, ARGV[2], ARGV[3])
    for _, c in ipairs(cont) do
        local k = cont_pfix .. c;
        local v = redis.call('HGET', k, 'objects')
        if v then
            v = tonumber(v)
            local s = tonumber(redis.call('HGET', k, 'bytes'))
            if v > tonumber(ARGV[1]) then
                res[c] = {v, s}
            end;
        end
    end;
    return cjson.encode(res);
`)

var scriptBucketInfo = redis.NewScript(`
	local buckets = redis.call("keys", "bucket:*");
	local res = {}
	for i, bucket in ipairs(buckets) do
      res[bucket] = {
          account = redis.call("hget", bucket, "account"),
          objects = tonumber(redis.call("hget", bucket, "objects")),
          bytes = tonumber(redis.call("hget", bucket, "bytes")),
      }
	end;
	return cjson.encode(res);`)

var scriptAcctInfo = redis.NewScript(`
	local accts = redis.call("hgetall", "accounts:")
	local res = {}
	local index = 0
	for i, acc in ipairs(accts) do
        if i % 2 == 1 then
                local acct_key = "account:" .. acc;
                res[index] = {acc, redis.call('HGET', acct_key, 'bytes'), redis.call("HGET", acct_key, 'objects')}
                index = index + 1
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

type bucketInfoStruct struct {
	Account string `json:"account"`
	Objects int64  `json:"objects"`
	Bytes   int64  `json:"bytes"`
}

// Collect -- collect container metrics
func Collect(client, bucketdb *redis.Client, ns string, l int64, t int64, f bool, c chan netdata.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if bucketdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		bucketInfoStr, err := scriptBucketInfo.Run(ctx, bucketdb, []string{}, 0).Result()
		if err != nil {
			return err
		}
		bucketInfo := map[string]bucketInfoStruct{}
		if err := json.Unmarshal([]byte(bucketInfoStr.(string)), &bucketInfo); err != nil {
			return err
		}
		for name, info := range bucketInfo {
			bucket := info.Account + "." + strings.Split(name, ":")[1]
			netdata.Update("account_bucket_kilobytes", bucket, fmt.Sprintf("%d", info.Bytes/1000), c)
			netdata.Update("account_bucket_objects", bucket, fmt.Sprintf("%d", info.Objects), c)
		}
	}

	accounts, err := scriptGetAccounts.Run(ctx, client, []string{}, 0).Result()
	if err != nil {
		return err
	}
	if f {
		acctInfo, err := scriptAcctInfo.Run(ctx, client, []string{}, 0).Result()
		if err != nil {
			return err
		}
		acctObj := map[string][]string{}
		err = json.Unmarshal([]byte(acctInfo.(string)), &acctObj)
		if err != nil {
			return err
		}
		for _, data := range acctObj {
			val, err := strconv.Atoi(data[1])
			if err != nil {
				return err
			}
			netdata.Update("account_bytes", util.AcctID(ns, data[0]), data[1], c)
			netdata.Update("account_kilobytes", util.AcctID(ns, data[0]), strconv.Itoa(val/1000), c)
			netdata.Update("account_objects", util.AcctID(ns, data[0]), data[2], c)
		}
	}

	for _, acct := range accounts.([]interface{}) {
		if acct == "1" {
			continue
		}
		count, err := scriptGetContCount.Run(ctx, client, []string{acct.(string)}, 1).Result()
		if err != nil {
			return err
		}
		ct := count.(int64)
		cts := strconv.FormatInt(ct, 10)
		netdata.Update("container_count", util.AcctID(ns, acct.(string)), cts, c)
		if !f {
			var i int64
			for i < ct {
				res, err := scriptListCont.Run(ctx, client, []string{acct.(string)}, t, i, l).Result()
				if err != nil {
					return err
				}
				contObj := map[string][]int{}
				err = json.Unmarshal([]byte(res.(string)), &contObj)
				if err != nil {
					return err
				}
				for cont, values := range contObj {
					netdata.Update("container_objects", util.AcctID(ns, acct.(string), cont), strconv.Itoa(values[0]), c)
					netdata.Update("container_bytes", util.AcctID(ns, acct.(string), cont), strconv.Itoa(values[1]), c)
				}
				i += l
				if l == -1 {
					i = ct
				}
			}
		}
	}
	return nil
}
