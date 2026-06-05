package statusz

// A circular buffer is created to hold the last N logs.
// The buffers are reallocated if needed, but otherwise reused to avoid
// allocations in the log Write path.

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

type lbuf struct {
	ln  int    // Length of log (not the buffer length)
	buf []byte // Allocated buffer
}

type memLogger struct {
	lock  sync.RWMutex
	size  uint // Number of log entries
	index uint // Index of next buffer to use
	cb    []lbuf
}

var memLog *memLogger // Lazily created memory logger
var logSetup sync.Once

// StdLoggerDefault attaches the memory logger to the default standard logger
func StdLoggerDefault(held uint) {
	StdLogger(log.Default(), held)
}

// StdLogger attaches the memory logger to a logger.
func StdLogger(logger *log.Logger, held uint) {
	logger.SetOutput(io.MultiWriter(MemLogger(held), logger.Writer()))
}

// MemLogger creates (if necessary) and provides a Writer that
// copies logs to the circular buffer.
func MemLogger(held uint) *memLogger {
	logSetup.Do(func() {
		memLog = new(memLogger)
		memLog.size = held
		memLog.cb = make([]lbuf, held)
		RegisterExtension(func(w http.ResponseWriter, r *http.Request) {
			memLog.logsExtension(w)
		})
	})
	return memLog
}

// logsExtension writes a HTML rendering of the logs.
func (l *memLogger) logsExtension(w http.ResponseWriter) {
	fmt.Fprint(w, "<h1>Recent logs</h1>")
	l.lock.RLock()
	defer l.lock.RUnlock()
	i := l.index // Oldest log
	for _ = range l.size {
		ent := &l.cb[i]
		if ent.ln != 0 {
			fmt.Fprintf(w, "%s<br>", string(ent.buf[:ent.ln]))
		}
		i = (i + 1) % l.size
	}
}

func (l *memLogger) Write(p []byte) (n int, err error) {
	ln := len(p)
	l.lock.Lock()
	defer l.lock.Unlock()
	ent := &l.cb[l.index]
	if ln > len(ent.buf) { // Check if buffer needs resizing
		ent.buf = make([]byte, ln)
	}
	// We don't resize the slice here, we just save the length separately.
	// That length can be used when the log is rendered.
	ent.ln = ln
	copy(ent.buf, p)
	l.index = (l.index + 1) % l.size
	return ln, nil
}
