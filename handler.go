// Package statusz implements a per-process status page showing runtime
// values and statistics. Custom handlers can be added to display
// extra information specific to the application
package statusz

import (
	"fmt"
    "html/template"
	"net/http"
	"os"
	"runtime/metrics"
	"sort"
	"strings"
	"time"
)

type handler func(http.ResponseWriter, *http.Request)

// link holds a URL path and the handler for that path
type link struct {
	path    string
	handler handler
}

// page holds a name of a page and the link to get to that page.
type page struct {
	name string
	path string
}

const (
	BasePage = "/statusz"
)

var (
	extensions []handler        // List of handlers to invoke for custom extras
	links      []link           // List of paths/handlers registered
	pages      []page           // List of linked pages to be added to /statusz
	muxes      []*http.ServeMux = []*http.ServeMux{http.DefaultServeMux}
)

type metricRef struct {
	label string
	name  string
}

// List of supported metrics at https://pkg.go.dev/runtime/metrics#hdr-Supported_metrics
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
	addLink(BasePage, statuszHandler)
}

// RegisterExtension adds a custom extension to the statusz page
// It is invoked after the standard statusz page
func RegisterExtension(f handler) {
	extensions = append(extensions, f)
}

// RegisterHandler registers a handler for a URL path under /statusz
func RegisterHandler(p string, handler handler) {
	addLink(BasePage+"/"+p, handler)
}

// RegisterPage adds a named link to the statusz page
func RegisterPage(name, path string, handler handler) {
	l := BasePage + "/" + path
	pages = append(pages, page{name: template.HTMLEscapeString(name), path: l})
	RegisterHandler(path, handler)
}

// Install adds statusz handling to this mux
func Install(mux *http.ServeMux) {
	// No need to install the default mux
	if mux == http.DefaultServeMux {
		return
	}
	// Add all of the handlers to this mux
	for _, l := range links {
		mux.HandleFunc(l.path, l.handler)
	}
	muxes = append(muxes, mux)
}

func addLink(path string, handler handler) {
	links = append(links, link{path: path, handler: handler})
	addHandler(path, handler)
}

// addHandler adds the path/handler to each of the muxes
func addHandler(path string, handler handler) {
	for _, m := range muxes {
		m.HandleFunc(path, handler)
	}
}

// statuszHandler implements the "/statusz" page
func statuszHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure that only "/statusz" is handled
	if r.URL.Path != BasePage {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<html><head></head><body>")
	fmt.Fprint(w, "<h1>Status</h1>")
	printRuntime(w)

	for _, f := range extensions {
		fmt.Fprint(w, "<hr>")
		f(w, r)
	}
	fmt.Fprint(w, "</body></html>")
}

// printRuntime displays the standard runtime information such as memory used,
// goroutines active etc.
func printRuntime(w http.ResponseWriter) {
	fmt.Fprintf(w, "Command line: %s<br>", template.HTMLEscapeString(strings.Join(os.Args, " ")))
	for _, p := range pages {
		fmt.Fprintf(w, "<a href=\"%s\">%s</a>, ", p.path, p.name)
	}
	fmt.Fprintf(w, "uptime %s", time.Since(startTime).Truncate(time.Second))
	if la, err := readProc("/proc/loadavg", 3); err == nil {
		fmt.Fprintf(w, ", Load avg: [%s]", strings.Join(la[:3], " "))
	} else {
		fmt.Fprintf(w, ", Load unavailable (%s)", err)
	}
	if st, err := readProc("/proc/self/stat", 22); err == nil {
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

// printMetricTable reads a set of associated metrics and displays them in a single table
func printMetricTable(w http.ResponseWriter, title, units string, names []metricRef) {
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

// readProc reads the selected file from "/proc" and parses the contents into
// a slice of strings.
func readProc(fn string, minFields int) ([]string, error) {
	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < minFields {
		return nil, fmt.Errorf("expected at least %d fields in %s - %d found", minFields, fn, len(fields))
	}
	return fields, nil
}

// format pretty-prints a metric value
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
		// May be seconds or cpu-seconds
		if strings.HasSuffix(unit, "seconds") {
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

// formatBytes pretty-prints a memory size value
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

func init() {
	RegisterPage("Environment", "environ", environHandler)
}

// environHandler is the page handler for displaying the environment variables for the process
func environHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("env").Parse(
`<html><head></head><body>
<h1>Environment variables</h1>
{{range .}}{{.}}<br>{{end}}
</body></html>`))
	w.Header().Set("Content-Type", "text/html")
	env := os.Environ()
	sort.Strings(env)
	tmpl.Execute(w, env)
}
