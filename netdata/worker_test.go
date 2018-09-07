package netdata

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

type testWriter struct {
	buf *bytes.Buffer
}

func (w *testWriter) Printf(format string, args ...interface{}) {
	w.buf.WriteString(fmt.Sprintf(format, args...))
}

func TestWorker(t *testing.T) {
	data := map[string]string{}
	collect := func() (map[string]string, error) {
		return data, nil
	}

	var buf bytes.Buffer
	writer := &testWriter{&buf}
	w := NewWorker(time.Millisecond, writer, collect)
	chart := NewChart("testType", "testID", "testName", "Test Title", "testUnit", "testFamily")
	chart.AddDimension("fooID", "foo", AbsoluteAlgorithm)
	chart.AddDimension("barID", "bar", IncrementalAlgorithm)
	w.AddChart(chart)

	expectedOutput := strings.Join([]string{
		"CHART testType.testID 'testName' 'Test Title' 'testUnit' 'testFamily'",
		"DIMENSION 'fooID' 'foo' absolute",
		"DIMENSION 'barID' 'bar' incremental",
	}, "\n")
	validateOutput(t, w, &buf, expectedOutput)

	// no output expected
	validateOutput(t, w, &buf, "")

	// dimension foo update
	data["fooID"] = "1"
	validateOutput(t, w, &buf, "BEGIN testType.testID\nSET 'fooID' = 1\nEND")

	// dimension bar update
	data["barID"] = "2"
	validateOutput(t, w, &buf, "BEGIN testType.testID\nSET 'fooID' = 1\nSET 'barID' = 2\nEND")

	delete(data, "fooID")
	validateOutput(t, w, &buf, "BEGIN testType.testID\nSET 'barID' = 2\nEND")

	delete(data, "barID")
	validateOutput(t, w, &buf, "")
}

func validateOutput(t *testing.T, w *worker, buf *bytes.Buffer, expectedOutput string) {
	w.process()
	output := buf.String()
	if output != expectedOutput {
		t.Fatalf("unexpected output got\n%q\nexpected\n%q\n", output, expectedOutput)
	}
	buf.Reset()
}
