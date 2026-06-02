package statusz

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type logBuffer struct {
	lock  sync.RWMutex
	size  uint
	index uint
	cb    []string
}

func Logs(held uint) {
	var logs logBuffer
	logs.size = held
	logs.cb = make([]string, held)
	log.SetOutput(io.MultiWriter(&logs, log.Writer()))
	RegisterExtension(func(w http.ResponseWriter, r *http.Request) {
		logs.logsExtension(w)
	})
}

// logs display
func (l *logBuffer) logsExtension(w http.ResponseWriter) {
	fmt.Fprint(w, "<h2>Recent logs</h2>")
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

func (l *logBuffer) Write(p []byte) (n int, err error) {
	var b strings.Builder
	ln, _ := b.Write(p)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.cb[l.index] = b.String()
	l.index = (l.index + 1) % l.size
	return ln, nil
}
