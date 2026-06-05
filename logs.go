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
	size  uint
	index uint
	cb    []lbuf
}

var memLog *memLogger
var logSetup sync.Once

func StdLoggerDefault(held uint) {
	StdLogger(log.Default(), held)
}

func StdLogger(logger *log.Logger, held uint) {
	logger.SetOutput(io.MultiWriter(MemLogger(held), logger.Writer()))
}

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

// logs display
func (l *memLogger) logsExtension(w http.ResponseWriter) {
	fmt.Fprint(w, "<h1>Recent logs</h1>")
	l.lock.RLock()
	defer l.lock.RUnlock()
	i := l.index
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
	// Check if buffer needs resizing
	if ln > len(ent.buf) {
		ent.buf = make([]byte, ln)
	}
	ent.ln = ln
	copy(ent.buf, p)
	l.index = (l.index + 1) % l.size
	return ln, nil
}
