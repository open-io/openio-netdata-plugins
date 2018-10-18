package netdata

import (
	"log"
	"time"
)

type Collector interface {
	Collect() (map[string]string, error)
}

type Worker struct {
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

func NewWorker(interval time.Duration, writer Writer, collector Collector) *Worker {
	return &Worker{
		interval:  interval,
		writer:    writer,
		collector: collector,
		charts:    make(map[string]*Chart),
	}
}

func (w *Worker) SetCollector(collector Collector) {
	w.collector = collector
}

func (w *Worker) AddChart(chart *Chart) {
	w.chartsIndex = append(w.chartsIndex, chart.ID)
	w.charts[chart.ID] = chart
}

func (w* Worker) SinceLastRun() time.Duration {
	return w.elapsed + w.interval;
}

func (w *Worker) Run() {
	log.Printf("Start interval: %v, retries: %v", w.interval, w.maxRetries)

	for {
		w.process()
	}
}

func (w *Worker) process() {
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

func (w *Worker) sleep(sleepTime time.Duration) {
	time.Sleep(sleepTime)
}

func (w *Worker) update(interval time.Duration) (bool, error) {
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
