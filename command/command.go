package command

import (
    "fmt"
    "os/exec"
    "oionetdata/netdata"
    "log"
    "strings"
    "time"
)

type Command struct {
    Desc string
    Family string
    Cmd string
}

type IWorker interface {
	SinceLastRun() time.Duration
	AddChart(chart *netdata.Chart)
}

type collector struct {
	cmds map[string]Command
    data map[string] string
    cache map[string]bool
    wait time.Duration
    cmdInterval time.Duration
    worker IWorker
}


func NewCollector(cmds map[string]Command, cmdInterval int64, w IWorker) *collector {
	return &collector{
		cmds: cmds,
        cmdInterval: time.Duration(cmdInterval)  * time.Second,
        data: nil,
        wait: time.Duration(0 * time.Second),
        cache: make(map[string]bool),
        worker: w,
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
    // TODO: Support multiple shells?
	out, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if err != nil {
        return "", err
	}
    outFmt := strings.TrimSuffix(string(out), "\n")
    // outFmt = strings.Replace(outFmt, ".", "_", -1)

	return outFmt, nil
}
