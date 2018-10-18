package main

import (
	"flag"
	"log"
	"os"
	"time"

	"oionetdata/collector"
	"oionetdata/netdata"
	"oionetdata/openio"
	"oionetdata/command"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}
	var conf string
	var cmdInterval int64
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.Int64Var(&cmdInterval, "interval", 3600, "Interval between commands in seconds")
	fs.StringVar(&conf, "conf", "/etc/netdata/commands.conf", "Command configuration file")
	fs.Parse(os.Args[2:])
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	cmds := make(map[string]command.Command)

	out, err := openio.Commands(conf)
	if err != nil {
		log.Fatalln("ERROR: Command plugin: Could not load commands", err)
	}

	for name, cmd := range out {
		cmds[name] = command.Command{Cmd: cmd, Desc: "OpenIO command", Family: "command"}
	}

	log.Printf("INFO: Command plugin: Loaded %d commands", len(cmds))

	writer := netdata.NewDefaultWriter()
	worker := netdata.NewWorker(time.Duration(intervalSeconds)*time.Second, writer, nil)
	collector := command.NewCollector(cmds, cmdInterval, worker)
	worker.SetCollector(collector)

	worker.Run()
}
