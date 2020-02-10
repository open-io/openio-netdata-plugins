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
	"log"
	"time"
)

type Collector interface {
	Collect() (map[string]string, error)
}

type worker struct {
	interval   time.Duration
	maxRetries int

	runs int

	startRun time.Time

	lastUpdate time.Time

	elapsed time.Duration

	charts      Charts
	chartsIndex map[Collector][]string

	writer Writer

	collector  Collector // Legacy attr, use collector
	collectors []Collector
}

func NewWorker(interval time.Duration, writer Writer, collectors ...Collector) *worker {
	w := worker{
		interval:    interval,
		writer:      writer,
		charts:      make(map[string]*Chart),
		chartsIndex: make(map[Collector][]string),
	}
	if len(collectors) > 0 {
		w.collector = collectors[0]
		w.collectors = append(w.collectors, collectors[0])
	}
	return &w
}

// Legacy method, use AddCollector
func (w *worker) SetCollector(collector Collector) {
	w.collector = collector
	w.collectors = append(w.collectors, collector)
}

func (w *worker) AddCollector(collector Collector) {
	w.collectors = append(w.collectors, collector)
}

func (w *worker) AddChart(chart *Chart, params ...Collector) {
	collector := w.collector
	if len(params) > 0 {
		collector = params[0]
	}
	chartID := fmt.Sprintf("%s_%s", chart.ID, chart.Family)
	w.indexChart(chartID, collector)
	w.charts[chartID] = chart
}

func (w *worker) GetChart(chartID string) *Chart {
	if chart, ok := w.charts[chartID]; ok {
		return chart
	}
	return nil
}

func (w *worker) indexChart(chartID string, collector Collector) {
	w.chartsIndex[collector] = append(w.chartsIndex[collector], chartID)
}

func (w *worker) SinceLastRun() time.Duration {
	return w.elapsed + w.interval
}

func (w *worker) Run() {
	log.Printf("Start interval: %v, retries: %v", w.interval, w.maxRetries)

	for {
		w.process()
	}
}

func (w *worker) process() {
	w.startRun = time.Now()

	var sinceUpdate time.Duration
	if !w.lastUpdate.IsZero() {
		sinceUpdate = w.startRun.Sub(w.lastUpdate)
	}
	updated, _ := w.update(sinceUpdate)

	w.runs++

	if !updated {
		// TODO manage retries?
	} else {
		w.elapsed = time.Since(w.startRun)
		w.lastUpdate = w.startRun
		log.Printf("elapsed: %v", w.elapsed)
	}

	w.sleep(w.interval)
}

func (w *worker) sleep(sleepTime time.Duration) {
	time.Sleep(sleepTime)
}

func (w *worker) update(interval time.Duration) (bool, error) {
	updated := false

	for _, collector := range w.collectors {
		data, err := collector.Collect()
		if err != nil {
			log.Printf("Failed to update: %v", err)
			continue
		}

		if _, ok := w.chartsIndex[collector]; ok {
			for _, chartID := range w.chartsIndex[collector] {
				chart := w.charts[chartID]
				updated = chart.Update(data, interval, w.writer)
			}
		} else {
			log.Printf("Failed to update: collector not found")
			log.Println(collector)
			log.Println("Charts index", w.chartsIndex)
		}

		if !updated {
			log.Printf("DEBUG: no charts updated")
		}
	}
	return updated, nil
}
