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

package command

import (
	"oionetdata/netdata"
	"oionetdata/util"
	"strings"
	"testing"
	"time"
)

type FakeWorker struct{}

func (w *FakeWorker) SinceLastRun() time.Duration {
	return 1000 * time.Second
}

func (w *FakeWorker) AddChart(chart *netdata.Chart, collector ...netdata.Collector) {
}

func keyExists(t *testing.T, data map[string]string, key string) {
	if _, ok := data[key]; !ok {
		t.Fatalf("Key %s not found in collected result data", key)
	}
}

func keyPrefixCount(t *testing.T, data map[string]string, prefix string, count int) {
	counter := 0
	for key := range data {
		if strings.HasPrefix(key, prefix) {
			counter++
		}
	}
	if counter != count {
		t.Fatalf("Expected %d occurrences for prefix %s, got %d: %v", count, prefix, counter, data)
	}
}

func valueCompare(t *testing.T, data map[string]string, key, value string, equal bool) {
	if _, ok := data[key]; !ok {
		t.Fatalf("Key %s not found in collected result data", key)
	}
	if equal && (data[key] != value) {
		t.Fatalf("Value of key %s should be %s", key, value)
	}
	if !equal && (data[key] == value) {
		t.Fatalf("Value of key %s should not be %s", key, value)
	}
}

func copyMap(src map[string]string) map[string]string {
	dst := make(map[string]string)
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func TestCommandCollector(t *testing.T) {
	cmds := util.Commands{Config: []util.Command{
		{Name: "test10", Command: "echo '1.2.3'", Family: "test"},
		{Name: "test20", Command: "date +%N", Family: "test"},
		{Name: "test21", Command: "date +%N", Family: "test", Interval: 10},
		{Name: "test30", Command: "echo '1.2'", Family: "test"},
		{Name: "test31", Command: "echo '1.2'", Family: "test", ValueIsLabel: true},
		{Name: "test40", Command: "echo -n 'v'; date +%N", Family: "test"},
	}}

	collector := NewCollector(cmds.Config, 1, &FakeWorker{})
	res2, err := collector.Collect()
	res := copyMap(res2)

	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)
	res2, err = collector.Collect()
	if err != nil {
		t.Fatal(err)
	}

	keyExists(t, res, "cmd_test10_1.2.3")
	keyExists(t, res, "cmd_test20")
	keyExists(t, res2, "cmd_test20")
	valueCompare(t, res2, "cmd_test20", res["cmd_test20"], false)
	keyExists(t, res, "cmd_test21")
	keyExists(t, res2, "cmd_test21")
	valueCompare(t, res2, "cmd_test21", res["cmd_test21"], true)
	keyExists(t, res, "cmd_test30")
	keyExists(t, res, "cmd_test31_1.2")
	keyPrefixCount(t, res, "cmd_test40", 1)
	keyPrefixCount(t, res2, "cmd_test40", 1)
}
