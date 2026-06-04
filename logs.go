package statusz

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type memLogger struct {
	lock  sync.RWMutex
	size  uint
	index uint
	cb    []string
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
		memLog.cb = make([]string, held)
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
		if l.cb[i] != "" {
			fmt.Fprintf(w, "%s<br>", l.cb[i])
		}
		i = (i + 1) % l.size
	}
}

func (l *memLogger) Write(p []byte) (n int, err error) {
	var b strings.Builder
	ln, _ := b.Write(p)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.cb[l.index] = b.String()
	l.index = (l.index + 1) % l.size
	return ln, nil
}
