package netdata

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

type testCollector struct {
	data map[string]string
}

func (c *testCollector) Collect() (map[string]string, error) {
	return c.data, nil
}

func TestWorker(t *testing.T) {
	data := map[string]string{}
	collector := &testCollector{data}
	var buf bytes.Buffer
	testWriter := &writer{out: &buf}
	w := NewWorker(time.Millisecond, testWriter, collector)
	chart := NewChart("testType", "testID", "testName", "Test Title", "testUnit", "testFamily", "testCategory")
	chart.AddDimension("fooID", "foo", AbsoluteAlgorithm)
	chart.AddDimension("barID", "bar", IncrementalAlgorithm)
	w.AddChart(chart)
	chart2 := NewChart("testType2", "testID2", "testName2", "Test Title 2", "testUnit2", "testFamily2", "testCategory2")
	chart2.AddDimension("foobarID", "foobar", AbsoluteAlgorithm)
	w.AddChart(chart2)

	expectedOutput := strings.Join([]string{
		"CHART testType.testID 'testName' 'Test Title' 'testUnit' 'testFamily' 'testCategory'",
		"DIMENSION 'fooID' 'foo' absolute",
		"DIMENSION 'barID' 'bar' incremental",
		"CHART testType2.testID2 'testName2' 'Test Title 2' 'testUnit2' 'testFamily2' 'testCategory2'",
		"DIMENSION 'foobarID' 'foobar' absolute",
		"",
	}, "\n")
	validateOutput(t, w, &buf, expectedOutput)

	// no output expected
	validateOutput(t, w, &buf, "")

	// dimension foo update
	data["fooID"] = "1"
	validateOutput(t, w, &buf, "BEGIN testType.testID\nSET 'fooID' = 1\nEND\n")

	// dimension bar update
	data["barID"] = "2"
	validateOutput(t, w, &buf, "BEGIN testType.testID\nSET 'fooID' = 1\nSET 'barID' = 2\nEND\n")

	delete(data, "fooID")
	validateOutput(t, w, &buf, "BEGIN testType.testID\nSET 'barID' = 2\nEND\n")

	delete(data, "barID")
	validateOutput(t, w, &buf, "")

	// dimension foo and foobar update
	data["fooID"] = "1"
	data["foobarID"] = "2"
	validateOutput(t, w, &buf, "BEGIN testType.testID\nSET 'fooID' = 1\nEND\nBEGIN testType2.testID2\nSET 'foobarID' = 2\nEND\n")

}

func validateOutput(t *testing.T, w *Worker, buf *bytes.Buffer, expectedOutput string) {
	w.process()
	output := buf.String()
	if output != expectedOutput {
		t.Fatalf("unexpected output got\n%q\nexpected\n%q\n", output, expectedOutput)
	}
	buf.Reset()
}
