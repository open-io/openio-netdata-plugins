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
	"fmt"
	"log"
	"oionetdata/netdata"
	"oionetdata/util"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Worker interface {
	SinceLastRun() time.Duration
	AddChart(chart *netdata.Chart, collector ...netdata.Collector)
}

type collector struct {
	cmds   []util.Command
	data   map[string]string
	cache  map[string]bool
	wait   time.Duration
	worker Worker
}

func setDefaults(cmds []util.Command, interval int64) {
	for i, c := range cmds {
		if c.Command == "" || c.Name == "" {
			log.Fatalf("Cannot parse command %d: fields name,command are required", i)
		}
		if c.Interval == 0 {
			c.Interval = interval
		}
		if c.Family == "" {
			c.Family = "command"
		}
		c.LastRun = 0
	}
}

func NewCollector(cmds []util.Command, interval int64, w Worker) *collector {
	setDefaults(cmds, interval)
	return &collector{
		cmds:   cmds,
		data:   make(map[string]string),
		wait:   time.Duration(0 * time.Second),
		cache:  make(map[string]bool),
		worker: w,
	}
}

func (c *collector) Collect() (map[string]string, error) {
	now := time.Now()

	for i, cmd := range c.cmds {
		if now.Unix()-cmd.LastRun < cmd.Interval {
			// Command is in cooldown
			continue
		}

		c.cmds[i].LastRun = now.Unix()

		value, err := c.runCommand(cmd.Command)

		if err != nil {
			log.Printf("WARN: Command collector: command %s failed with error %s", cmd.Command, err)
			continue
		}
		if value == "" {
			log.Printf("WARN: Command collector: command %s returned no output", cmd.Command)
			continue
		}

		chart := fmt.Sprintf("cmd_%s", cmd.Name)
		valueAsLabel := false

		// Check if the value can be reported as is
		if _, err := strconv.ParseFloat(value, 64); cmd.ValueIsLabel || err != nil {
			chart = fmt.Sprintf("cmd_%s_%v", cmd.Name, value)
			valueAsLabel = true
		}

		// Check if a new chart needs to be created
		if _, ok := c.cache[chart]; !ok {
			newChart := netdata.NewChart(chart, cmd.Name, "", cmd.Name, "", cmd.Family, "command")
			newChart.AddDimension(chart, cmd.Name, netdata.AbsoluteAlgorithm)
			c.worker.AddChart(newChart)
			c.cache[chart] = true
		}
		if cmd.ValueIsLabel || valueAsLabel {
			value = fmt.Sprintf("%d", now.Unix())
		}
		c.data[chart] = value
	}

	return c.data, nil
}

func (c *collector) runCommand(cmd string) (string, error) {
	out, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	outFmt := strings.Trim(
		strings.TrimSuffix(string(out), "\n"), " ")

	return outFmt, nil
}
