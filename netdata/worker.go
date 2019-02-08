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
	chartsIndex []string

	writer Writer

	collector Collector
}

func NewWorker(interval time.Duration, writer Writer, collector Collector) *worker {
	return &worker{
		interval:  interval,
		writer:    writer,
		collector: collector,
		charts:    make(map[string]*Chart),
	}
}

func (w *worker) SetCollector(collector Collector) {
	w.collector = collector
}

func (w *worker) AddChart(chart *Chart) {
	w.chartsIndex = append(w.chartsIndex, chart.ID)
	w.charts[chart.ID] = chart
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
	sleepTime := w.interval

	w.sleep(sleepTime)

	w.startRun = time.Now()

	var sinceUpdate time.Duration
	if !w.lastUpdate.IsZero() {
		sinceUpdate = w.startRun.Sub(w.lastUpdate)
	}
	updated, err := w.update(sinceUpdate)
	if err != nil {
		log.Printf("Failed to update: %v", err)
	}

	w.runs++

	if !updated {
		// TODO manage retries?
	} else {
		w.elapsed = time.Since(w.startRun)
		w.lastUpdate = w.startRun
		log.Printf("elapsed: %v", w.elapsed)
	}
}

func (w *worker) sleep(sleepTime time.Duration) {
	time.Sleep(sleepTime)
}

func (w *worker) update(interval time.Duration) (bool, error) {
	data, err := w.collector.Collect()
	if err != nil {
		return false, err
	}

	updated := false

	for _, chartID := range w.chartsIndex {
		chart := w.charts[chartID]
		updated = chart.Update(data, interval, w.writer)
	}

	if !updated {
		log.Printf("DEBUG: no charts updated")
	}

	return updated, nil
}
