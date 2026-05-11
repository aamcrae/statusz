package statusz

import (
	"fmt"
	"net/http"
	"os"
	"runtime/metrics"
	"sort"
	"strings"
	"time"
)

type localHandler func(http.ResponseWriter, *http.Request)

var localHandlers []localHandler

const (
	BasePage       = "/statusz"
	environ        = "environ"
	flightRecorder = "recorder"
)

type metricRef struct {
	label string
	name  string
}

var cpuMetrics = []metricRef{
	{"GC total time", "/cpu/classes/gc/total:cpu-seconds"},
	{"CPU time used", "/cpu/classes/user:cpu-seconds"},
	{"CPU idle time", "/cpu/classes/idle:cpu-seconds"},
	{"CPU total time", "/cpu/classes/total:cpu-seconds"},
}

var schedulerMetrics = []metricRef{
	{"OS threads available", "/sched/gomaxprocs:threads"},
	{"Goroutines created", "/sched/goroutines-created:goroutines"},
	{"Live goroutines", "/sched/goroutines:goroutines"},
	{"Runnable goroutines", "/sched/goroutines/runnable:goroutines"},
	{"Running goroutines", "/sched/goroutines/running:goroutines"},
	{"Waiting goroutines", "/sched/goroutines/waiting:goroutines"},
	{"Seconds blocked on lock", "/sync/mutex/wait/total:seconds"},
}

var memMetrics = []metricRef{
	{"Total memory used", "/memory/classes/total:bytes"},
	{"Heap objects", "/memory/classes/heap/objects:bytes"},
	{"Heap stacks", "/memory/classes/heap/stacks:bytes"},
	{"Heap free", "/memory/classes/heap/free:bytes"},
	{"Heap unused", "/memory/classes/heap/unused:bytes"},
	{"Heap released", "/memory/classes/heap/released:bytes"},
}

// For uptime
var startTime = time.Now()

func init() {
	http.HandleFunc(BasePage, statuszHandler)
	RegisterPage(environ, environHandler)
	RegisterPage(flightRecorder, flightRecorderHandler)
}

func RegisterLocalHandler(f localHandler) {
	localHandlers = append(localHandlers, f)
}

func RegisterPage(p string, f func(http.ResponseWriter, *http.Request)) {
	if p != "" {
		http.HandleFunc(BasePage+"/"+p, f)
	}
}

func statuszHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure that there are no extra references
	if r.URL.Path != BasePage {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<html><head></head><body>")
	fmt.Fprint(w, "<h1>Status</h1>")
	printRuntime(w)

	for _, f := range localHandlers {
		fmt.Fprint(w, "<hr>")
		f(w, r)
	}
}

func printRuntime(w http.ResponseWriter) {
	fmt.Fprintf(w, "Command line: %s<br>", strings.Join(os.Args, " "))
	fmt.Fprint(w, "<a href=\""+BasePage+"/"+environ+"\">Environment</a>, ")
	fmt.Fprint(w, "<a href=\""+BasePage+"/"+flightRecorder+"\">Flight Recorder</a>, ")
	fmt.Fprintf(w, "uptime %s", time.Since(startTime).Truncate(time.Second))
	if la, err := readProc("/proc/loadavg", " ", 3); err == nil {
		fmt.Fprintf(w, ", Load avg: [%s]", strings.Join(la[:3], " "))
	} else {
		fmt.Fprintf(w, ", Load unavailable (%s)", err)
	}
	if st, err := readProc("/proc/self/stat", " ", 22); err == nil {
		fmt.Fprintf(w, ", PID %s", st[0])
	} else {
		fmt.Fprintf(w, ", Process information unavailable (%s)", err)
	}
	fmt.Fprint(w, "<p><div style=\"display:flex; justify-content: space-evenly;\">")
	printMetricTable(w, "CPU time", "CPU Seconds", cpuMetrics)
	printMetricTable(w, "Scheduler", "Count or seconds", schedulerMetrics)
	printMetricTable(w, "Memory", "Size", memMetrics)
	fmt.Fprintf(w, "</div>")
}

func printMetricTable(w http.ResponseWriter, title string, units string, names []metricRef) {
	samples := make([]metrics.Sample, len(names))
	for i := range names {
		samples[i].Name = names[i].name
	}
	metrics.Read(samples)
	fmt.Fprintf(w, "<table border=\"1\"><thead><tr><th>%s</th><th>%s</th></tr></thead><tbody>", title, units)
	for i, sample := range samples {
		fmt.Fprintf(w, "<tr><td>%s</td><td style=\"text-align: end\">%s</td></tr>", names[i].label, format(sample))
	}
	fmt.Fprint(w, "</tbody></table>")
}

func readProc(fn string, sep string, minFields int) ([]string, error) {
	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	fields := strings.Split(string(data), sep)
	if len(fields) < minFields {
		return nil, fmt.Errorf("expected at least %d fields in %s - %d found", minFields, fn, len(fields))
	}
	return fields, nil
}

func format(s metrics.Sample) string {
	// Extract unit
	_, unit, _ := strings.Cut(s.Name, ":")
	switch s.Value.Kind() {
	case metrics.KindUint64:
		if unit == "bytes" {
			return formatBytes(s.Value.Uint64())
		} else {
			return fmt.Sprintf("%d", s.Value.Uint64())
		}
	case metrics.KindFloat64:
		if unit == "cpu-seconds" {
			return fmt.Sprintf("%.3f", s.Value.Float64())
		} else if unit == "seconds" {
			return fmt.Sprintf("%.3f", s.Value.Float64())
		} else {
			return fmt.Sprintf("%f", s.Value.Float64())
		}
	case metrics.KindBad:
		return "unknown metric"
	default:
		return "unexpected metric type"
	}
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func environHandler(w http.ResponseWriter, _ *http.Request) {
	env := os.Environ()
	sort.Strings(env)
	fmt.Fprint(w, strings.Join(env, "\n"))
}
