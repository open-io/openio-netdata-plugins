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

package netdata

import (
	"fmt"
	"strings"
	"sync"
)

type index struct {
	sync.RWMutex
	charts map[string]map[string]bool
}

func makeIndex() *index {
	return &index{
		charts: make(map[string]map[string]bool),
	}
}

func (i *index) chartExists(chart string) bool {
	i.RLock()
	defer i.RUnlock()
	_, e := i.charts[chart]
	return e
}

func (i *index) dimExists(chart string, dim string) bool {
	i.RLock()
	defer i.RUnlock()
	_, e := i.charts[chart][dim]
	return e
}

func (i *index) addChart(chart string) {
	i.Lock()
	defer i.Unlock()
	i.charts[chart] = make(map[string]bool)
}

func (i *index) addDim(chart string, dim string) {
	i.Lock()
	defer i.Unlock()
	i.charts[chart][dim] = true
}

var chartIndex = makeIndex()

// Prefix -- prefix to use for metrics
var Prefix = "openio"

/*
Metric - metric to be sent to buffer
*/
type Metric struct {
	Chart string
	Dim   string
	Value string
}

/*
Update - queue a new metric value on a chart
*/
func Update(chart string, dim string, value string, c chan Metric) {
	chart = fmt.Sprintf("%s.%s", Prefix, strings.Replace(chart, ".", "_", -1))
	chartTitle := strings.ToUpper(strings.Join(strings.Split(chart, "_"), " "))
	if !chartIndex.chartExists(chart) {
		createChart(chart, "", chartTitle, "", "")
		chartIndex.addChart(chart)
	}
	if !chartIndex.dimExists(chart, dim) {
		createChart(chart, "", chartTitle, "", dim)
		chartIndex.addDim(chart, dim)
	}

	c <- Metric{
		Chart: chart,
		Dim:   dim,
		Value: value,
	}
}

func createChart(chart string, desc string, title string, units string, dim string) {
	if dim != "" {
		dim = fmt.Sprintf("DIMENSION %s '%s' absolute\n", dim, dim)
	}
	fmt.Printf("CHART %s '%s' '%s' '%s' '%s'\n%s", chart, desc, title, units, getFamily(chart), dim)
}

func getFamily(chart string) string {
	families := map[string]string{
		"req":       "Request",
		"rep":       "Response",
		"score":     "Score",
		"byte":      "Capacity",
		"inodes":    "Inodes",
		"cnx":       "Connections",
		"zk":        "Zookeeper",
		"container": "Container",
		"account":   "Account",
	}

	chart = strings.Split(chart, ".")[1]
	for k, v := range families {
		if strings.HasPrefix(chart, k) {
			return v
		}
	}
	return "Misc"
}
