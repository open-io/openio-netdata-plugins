package collector

import (
	"time"
	"oionetdata/netdata"
	"fmt"
    "strconv"
)

// PollsBeforeReload -- Number of cycles before collector is reloaded
// Restarting the collector every now and then should help getting rid of memleaks
var PollsBeforeReload = 1000;

var buf = make(map[string][]byte)

// Collect -- function to call on each collection
type Collect func(chan netdata.Metric)


// ParseInterval -- parse interval from arguments
func ParseInterval(args []string) ([]string, int64) {
    var interval int64;
    var err error;
	if len(args) > 1 {
		interval, err = strconv.ParseInt(args[1], 10, 0)
		if err != nil {
            interval = 10
        } else {
            args = append(args[:1], args[2:]...)
        }
	}
    return args, interval
}

// Run -- run the collector
func Run(pollInt int64, collect Collect) {
    poll := 0

	for poll < PollsBeforeReload {
		c := make(chan netdata.Metric, 1e5)
		collect(c)
		time.Sleep(time.Duration(pollInt) * 1000 * time.Millisecond);
		close(c)
		for m := range c {
			buf[m.Chart] = append(buf[m.Chart], fmt.Sprintf("SET %s %s\n", m.Dim, m.Value)...)
		}
		for c, v := range buf {
			fmt.Printf("BEGIN %s\n%sEND\n", c, string(v))
		}
		poll++;
        // Send & reset the buffer after the collection
		buf = make(map[string][]byte)
	}
}
