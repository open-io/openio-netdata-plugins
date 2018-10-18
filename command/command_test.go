package command

import (
	"oionetdata/netdata"
	"testing"
	"time"
)

type FakeWorker struct{}

func (w *FakeWorker) SinceLastRun() time.Duration {
	return 1000 * time.Second
}

func (w *FakeWorker) AddChart(chart *netdata.Chart) {
	return
}

func TestCommandCollector(t *testing.T) {
	cmds := make(map[string]Command)
	cmds["test1"] = Command{Cmd: "echo '1.2.3'", Desc: "test", Family: "test"}
	cmds["test2"] = Command{Cmd: "echo '1.2.4'", Desc: "test", Family: "test"}

	collector := NewCollector(cmds, 10, &FakeWorker{})
	res, err := collector.Collect()

	if err != nil {
		t.Fatal(err)
	}

	testData := map[string]string{
		"cmd_test1_1.2.3": "1",
		"cmd_test2_1.2.4": "1",
	}

	// Test returned data
	for k := range res {
		if _, ok := testData[k]; !ok {
			t.Fatalf("Key %s not found in collected result data", k)
		}
	}
}
