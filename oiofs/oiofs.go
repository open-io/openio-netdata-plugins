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
)

type collector struct {
	addr string
}

func NewCollector(addr string) *collector {
	return &collector{
		addr: addr,
	}
}

func (c *collector) Collect() (map[string]string, error) {
	// TODO: support v2
	r, err := http.Get(fmt.Sprintf("http://%s/stats", c.addr))
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

	res := map[string]string{}

	for k, v := range(rs.(map[string]interface{})) {
		res[k] = strconv.FormatInt(int64(v.(float64)), 10)
	}
	return res, nil
}
