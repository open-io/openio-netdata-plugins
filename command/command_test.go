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
	"testing"
	"time"
)

type FakeWorker struct{}

func (w *FakeWorker) SinceLastRun() time.Duration {
	return 1000 * time.Second
}

func (w *FakeWorker) AddChart(chart *netdata.Chart, collector ...netdata.Collector) {
}

func TestCommandCollector(t *testing.T) {
	cmds := make(map[string]Command)
	cmds["test1"] = Command{Cmd: "echo '1.2.3'", Desc: "test", Family: "test"}
	cmds["test2"] = Command{Cmd: "echo '1.2.4'", Desc: "test", Family: "test"}
	cmds["test3"] = Command{Cmd: "echo '1.2.5 '", Desc: "test", Family: "test"}

	collector := NewCollector(cmds, 10, &FakeWorker{})
	res, err := collector.Collect()

	if err != nil {
		t.Fatal(err)
	}

	testData := map[string]string{
		"cmd_test1_1.2.3": "1",
		"cmd_test2_1.2.4": "1",
		"cmd_test3_1.2.5": "1",
	}

	// Test returned data
	for k := range res {
		if _, ok := testData[k]; !ok {
			t.Fatalf("Key %s not found in collected result data", k)
		}
	}
}
