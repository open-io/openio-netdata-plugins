package container

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"oionetdata/netdata"
	"oionetdata/util"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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

// RedisAddr -- get redis address
func RedisAddr(basePath string, ns string) string {
	ip := ""
	port := ""
	p := path.Join(basePath, ns, "redis-*/redis.conf")
	match, err := filepath.Glob(p)
	if err != nil || len(match) == 0 {
		util.RaiseIf(fmt.Errorf("ERROR Path %s did not match any files", p))
	}

	file, err := os.Open(match[0])
	util.RaiseIf(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t := scanner.Text()
		if strings.HasPrefix(t, "port ") {
			port = strings.Split(t, "port ")[1]
		}
		if strings.HasPrefix(t, "bind ") {
			ip = strings.Split(t, "bind ")[1]
		}
		if ip != "" && port != "" {
			return fmt.Sprintf("%s:%s", ip, port)
		}
	}
	return ""
}

// Collect -- collect container metrics
func Collect(client *redis.Client, ns string, l int64, t int64, c chan netdata.Metric) error {

	accounts, err := scriptGetAccounts.Run(client, []string{}, 0).Result()
	if err != nil {
		return err
	}

	for _, acct := range accounts.([]interface{}) {
		if acct == "1" {
			continue
		}
		count, err := scriptGetContCount.Run(client, []string{acct.(string)}, 1).Result()
		if err != nil {
			return err
		}
		ct := count.(int64)
		cts := strconv.FormatInt(ct, 10)
		netdata.Update("container_count", util.AcctID(ns, acct.(string)), cts, c)
		var i int64
		for i < ct {
			res, err := scriptListCont.Run(client, []string{acct.(string)}, t, i, l).Result()
			if err != nil {
				return err
			}
			contObj := map[string][]int{}
			err = json.Unmarshal([]byte(res.(string)), &contObj)
			util.RaiseIf(err)
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
	return nil
}
