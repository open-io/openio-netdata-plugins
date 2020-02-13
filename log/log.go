package logger

import (
	"fmt"
	"github.com/hpcloud/tail"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PerfData struct {
	bwIn    int64
	bwOut   int64
	rt      float64
	rtMin   float64
	rtMax   float64
	rtCount int64
}

type metrics struct {
	sync.Mutex
	Requests map[string]int64
	PerfData map[string]*PerfData
}

type collector struct {
	retry       time.Duration
	lastCollect time.Time
	metrics     *metrics
}

func NewCollector(path string) *collector {
	retry, targets, err := ParseLogConfig(path)
	if err != nil {
		log.Fatalln("ERROR: Log collector: Could not parse configuration", err)
	}

	c := &collector{
		retry:       retry,
		lastCollect: time.Now(),
		metrics: &metrics{
			Requests: map[string]int64{},
			PerfData: map[string]*PerfData{},
		},
	}

	for name, t := range targets {
		go c.tailLogFile(name, t)
	}

	return c
}

type LogTarget struct {
	Name   string `yaml:"name"`
	Path   string `yaml:"path"`
	Format struct {
		Pattern string `yaml:"pattern"`
	} `yaml:"custom_log_format"`
	Filter struct {
		Include string `yaml:"include"`
		Exclude string `yaml:"exclude"`
	}
}

var tailConfig = tail.Config{
	// Logger: tail.DiscardingLogger,
	Follow: true,
	// This prevents channel closure on file moves
	ReOpen: true,
	// This prevents tail from re-parsing previous lines
	Location: &tail.SeekInfo{Offset: 0, Whence: 2},
}

func ParseLogConfig(path string) (retry time.Duration, targets map[string]*LogTarget, err error) {
	cast := make(map[string]interface{})

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &cast)
	if err != nil {
		return
	}

	retry = 60 * time.Second

	if v, ok := cast["autodetection_retry"]; ok {
		if v2, ok := v.(int); ok {
			retry = time.Duration(v2) * time.Second
		}
	}

	// Re-marshal a filtered map containing only valid targets,
	// then unmarshal it back into the proper struct
	// Note: maybe rewrite using reflect
	filtered, err := yaml.Marshal(func() map[string]interface{} {
		res := map[string]interface{}{}
		for k, v := range cast {
			if _, ok := v.(string); ok {
				continue
			}
			if _, ok := v.(int64); ok {
				continue
			}
			if _, ok := v.(float64); ok {
				continue
			}
			res[k] = v
		}
		return res
	}())
	if err != nil {
		return
	}
	if err2 := yaml.Unmarshal(filtered, &targets); err2 != nil {
		return
	}

	return
}

func (c *collector) Collect() (map[string]string, error) {
	duration := time.Since(c.lastCollect).Seconds()
	c.lastCollect = time.Now()
	data := map[string]string{}

	c.metrics.Lock()
	for k, v := range c.metrics.Requests {
		data["response_code_log_ops_"+k] = fmt.Sprintf("%g", float64(v)/duration)
	}
	for k, v := range c.metrics.PerfData {
		data["response_time_min_"+k] = fmt.Sprintf("%g", v.rtMin)
		data["response_time_max_"+k] = fmt.Sprintf("%g", v.rtMax)
		if v.rtCount > 0 {
			data["response_time_avg_"+k] = fmt.Sprintf("%g", v.rt/float64(v.rtCount))
		} else {
			data["response_time_avg_"+k] = "0"
		}
		data["bandwidth_in_"+k] = fmt.Sprintf("%.0f", float64(v.bwIn/1024)/duration)
		data["bandwidth_out_"+k] = fmt.Sprintf("%.0f", float64(v.bwOut/1024)/duration)
	}



	for k := range c.metrics.Requests {
		c.metrics.Requests[k] = 0
	}
	for k := range c.metrics.PerfData {
		c.metrics.PerfData[k] = &PerfData{
			rtMax:   0,
			rtMin:   0,
			rtCount: 0,
			rt:      0,
		}
	}
	c.metrics.Unlock()
	return data, nil
}

func tagString(tags []string, matches []string) string {
	res := []string{}
	for i, tag := range tags {
		if strings.HasPrefix(tag, "tag_") {
			res = append(res, matches[i])
		}
	}
	rs := strings.Join(res, ".")
	if rs != "" {
		rs = "." + rs
	}
	return rs
}

func index(data []string, key string) int {
	for i, v := range data {
		if key == v {
			return i
		}
	}
	return -1
}

func (c *collector) tailLogFile(name string, conf *LogTarget) {
	rTokenize, err := regexp.Compile(conf.Format.Pattern)
	if err != nil {
		log.Println("WARN: Fatal error watching file", conf.Path, ":", err)
		return
	}
	numTokens := len(rTokenize.SubexpNames())
	var rInclude *regexp.Regexp
	var rExclude *regexp.Regexp

	if conf.Filter.Include != "" {
		rInclude, err = regexp.Compile(conf.Filter.Include)
		if err != nil {
			log.Println("WARN: Fatal error watching file", conf.Path, ":", err)
			return
		}
	}
	if conf.Filter.Exclude != "" {
		rExclude, err = regexp.Compile(conf.Filter.Exclude)
		if err != nil {
			log.Println("WARN: Fatal error watching file", conf.Path, ":", err)
			return
		}
	}

	for {
		if _, err := os.Stat(conf.Path); err != nil {
			log.Println("WARN: Could not watch file", conf.Path, ":", err, ". Retrying in", c.retry)
			time.Sleep(c.retry * time.Second)
			continue
		}
		t, err := tail.TailFile(conf.Path, tailConfig)
		if err != nil {
			log.Println("WARN: Error watching file", conf.Path, ":", err)
			return
		}

		for line := range t.Lines {
			if line.Err != nil {
				log.Println("WARN: Error watching file", conf.Path, ":", err)
				continue
			}
			if rInclude != nil && !rInclude.MatchString(line.Text) {
				continue
			}
			if rExclude != nil && rExclude.MatchString(line.Text) {
				continue
			}
			s := rTokenize.FindStringSubmatch(line.Text)
			if len(s) != numTokens {
				continue
			}
			names := rTokenize.SubexpNames()
			tagString := tagString(names, s)

			codeIdx := index(names, "code")
			method := index(names, "method")
			pfx := s[method] + "."
			if codeIdx > -1 {
				reqName := pfx + s[codeIdx] + "." + name + tagString
				// NOTE: consider commit at the end to avoid locking N times
				c.metrics.Lock()
				if _, ok := c.metrics.Requests[reqName]; ok {
					c.metrics.Requests[reqName]++
				} else {
					c.metrics.Requests[reqName] = 1
				}
				c.metrics.Unlock()
			}

			reqName := pfx + name + tagString

			bsIdx := index(names, "bytes_sent")
			if bsIdx > -1 {
				vInt64, err := strconv.ParseInt(s[bsIdx], 10, 64)
				if err == nil {
					c.metrics.Lock()
					if _, ok := c.metrics.PerfData[reqName]; ok {
						c.metrics.PerfData[reqName].bwOut += vInt64
					} else {
						c.metrics.PerfData[reqName] = &PerfData{
							rt:      0,
							rtMin:   0,
							rtMax:   0,
							rtCount: 0,
							bwIn:    0,
							bwOut:   vInt64,
						}
					}
					c.metrics.Unlock()
				}
			}
			brIdx := index(names, "resp_length")
			if brIdx > -1 {
				vInt64, err := strconv.ParseInt(s[brIdx], 10, 64)
				if err == nil {
					c.metrics.Lock()
					if _, ok := c.metrics.PerfData[reqName]; ok {
						c.metrics.PerfData[reqName].bwIn += vInt64
					} else {
						c.metrics.PerfData[reqName] = &PerfData{
							rt:      0,
							rtMin:   0,
							rtMax:   0,
							rtCount: 0,
							bwIn:    vInt64,
							bwOut:   0,
						}
					}
					c.metrics.Unlock()
				}
			}
			rtIdx := index(names, "resp_time")
			if rtIdx > -1 {
				vFloat, err := strconv.ParseFloat(s[rtIdx], 4)
				if err == nil {
					c.metrics.Lock()
					if _, ok := c.metrics.PerfData[reqName]; ok {
						if vFloat < c.metrics.PerfData[reqName].rtMin {
							c.metrics.PerfData[reqName].rtMin = vFloat
						} else if vFloat > c.metrics.PerfData[reqName].rtMax {
							c.metrics.PerfData[reqName].rtMax = vFloat
						}
						c.metrics.PerfData[reqName].rt += vFloat
					} else {
						c.metrics.PerfData[reqName] = &PerfData{
							rt:      vFloat,
							rtMin:   vFloat,
							rtMax:   vFloat,
							rtCount: 1,
							bwIn:    0,
							bwOut:   0,
						}
					}
					c.metrics.Unlock()
				}
			}
		}
	}
}
