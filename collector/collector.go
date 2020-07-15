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

package collector

import (
	"fmt"
	"log"
	"github.com/open-io/openio-netdata-plugins/netdata"
	"strconv"
	"time"
)

// PollsBeforeReload -- Number of cycles before collector is reloaded
// Restarting the collector every now and then should help getting rid of memleaks
var PollsBeforeReload = 1000

var buf = make(map[string][]byte)

// Collect -- function to call on each collection
type Collect func(chan netdata.Metric) error

const DefaultIntervalSeconds = 10

// ParseIntervalSeconds parses the interval
func ParseIntervalSeconds(arg string) int {
	interval, err := strconv.Atoi(arg)
	if err != nil {
		return DefaultIntervalSeconds
	}
	return interval
}

// Run -- run the collector
func Run(intervalSeconds int, collect Collect) {
	poll := 0
	cd := 1
	maxCd := intervalSeconds * 20

	for poll < PollsBeforeReload {
		c := make(chan netdata.Metric, 1e5)
		err := collect(c)
		if err != nil {
			cd += cd
			if cd > maxCd {
				cd = maxCd
			}
			log.Println("Collect function returned an error", err)
			log.Printf("Retrying collection in %d second(s)", cd)
			time.Sleep(time.Duration(cd) * time.Second)
		} else {
			cd = 1
		}
		time.Sleep(time.Duration(intervalSeconds) * time.Second)
		close(c)
		for m := range c {
			buf[m.Chart] = append(buf[m.Chart], fmt.Sprintf("SET %s %s\n", m.Dim, m.Value)...)
		}
		for c, v := range buf {
			fmt.Printf("BEGIN %s\n%sEND\n", c, string(v))
		}
		poll++
		// Send & reset the buffer after the collection
		buf = make(map[string][]byte)
	}
}
