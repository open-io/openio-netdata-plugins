package netdata

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Writer interface {
	Printf(format string, v ...interface{})
}

func NewDefaultWriter() Writer {
	return &writer{
		out: os.Stdout,
	}
}

type writer struct {
	sync.Mutex
	out io.Writer
}

func (w *writer) Printf(format string, v ...interface{}) {
	w.Lock()
	fmt.Fprintf(w.out, format, v...)
	w.Unlock()
}
