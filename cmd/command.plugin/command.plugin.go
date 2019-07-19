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

package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"oionetdata/collector"
	"oionetdata/command"
	"oionetdata/netdata"
	"oionetdata/util"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}
	var conf string
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&conf, "conf", "/etc/netdata/commands.conf", "Command configuration file")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		log.Fatalln("ERROR: Command plugin: Could not parse args", err)
	}
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	cmds := util.Commands{}

	if strings.HasSuffix(conf, ".yml") || strings.HasSuffix(conf, ".yaml") {
		cmds, err = util.ParseCommandsYaml(conf)
	} else {
		cmds, err = util.ParseCommands(conf)
	}
	if err != nil {
		log.Fatalln("ERROR: Command plugin: Could not load commands", err)
	}

	log.Printf("INFO: Command plugin: Loaded %d commands from %s", len(cmds.Config), conf)

	writer := netdata.NewDefaultWriter()
	worker := netdata.NewWorker(time.Duration(intervalSeconds)*time.Second, writer)
	collector := command.NewCollector(cmds.Config, int64(intervalSeconds), worker)
	worker.SetCollector(collector)

	worker.Run()
}
