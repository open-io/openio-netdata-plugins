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
	"github.com/open-io/openio-netdata-plugins/netdata"
	"os/exec"
	"strings"
	"time"
)

type Command struct {
	Desc   string
	Family string
	Cmd    string
}

type Worker interface {
	SinceLastRun() time.Duration
	AddChart(chart *netdata.Chart, collector ...netdata.Collector)
}

type collector struct {
	cmds        map[string]Command
	data        map[string]string
	cache       map[string]bool
	wait        time.Duration
	cmdInterval time.Duration
	worker      Worker
}

func NewCollector(cmds map[string]Command, cmdInterval int64, w Worker) *collector {
	return &collector{
		cmds:        cmds,
		cmdInterval: time.Duration(cmdInterval) * time.Second,
		data:        nil,
		wait:        time.Duration(0 * time.Second),
		cache:       make(map[string]bool),
		worker:      w,
	}
}

func (c *collector) Collect() (map[string]string, error) {

	c.wait -= c.worker.SinceLastRun()

	if c.wait > 0 {
		return c.data, nil
	}

	c.wait = c.cmdInterval

	c.data = make(map[string]string)

	for name, op := range c.cmds {
		value, err := c.runCommand(op.Cmd)
		if err != nil {
			log.Printf("WARN: Command collector: command %s failed with error %s", op.Cmd, err)
			continue
		}

		chart := fmt.Sprintf("cmd_%s_%s", name, value)
		if _, ok := c.cache[chart]; !ok {
			newChart := netdata.NewChart(chart, name, "", op.Desc, "", op.Family, "command")
			newChart.AddDimension(chart, name, netdata.AbsoluteAlgorithm)
			c.worker.AddChart(newChart)
			c.cache[chart] = true
		}

		c.data[chart] = fmt.Sprintf("%d", time.Now().Unix())
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
