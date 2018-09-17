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

func (w *worker) AddChart(chart *Chart) {
	w.chartsIndex = append(w.chartsIndex, chart.ID)
	w.charts[chart.ID] = chart
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
		elapsed := time.Since(w.startRun)
		w.lastUpdate = w.startRun
		log.Printf("elapsed: %v", elapsed)
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
