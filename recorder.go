package statusz

import (
	"fmt"
	"net/http"
	"runtime/trace"
	"strconv"
	"sync"
	"time"
)

var recorderLock sync.Mutex
var recorderTime int = 10 // seconds
var recorderSize int = 5  // Mbyte
var recorder *trace.FlightRecorder

func init() {
	RegisterPage(flightRecorder+"/start", flightRecorderStart)
	RegisterPage(flightRecorder+"/stop", flightRecorderStop)
	RegisterPage(flightRecorder+"/download", flightRecorderDownload)
}

func flightRecorderHandler(w http.ResponseWriter, r *http.Request) {
	var st string
	recorderLock.Lock()
	defer recorderLock.Unlock()
	if recorder != nil {
		st = "RUNNING"
	} else {
		st = "STOPPED"
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<html><head></head><body>")
	fmt.Fprint(w, "<h1>Flight Recorder</h1>")
	fmt.Fprintf(w, "<iframe name=\"response\" srcdoc=\"%s\" height=\"44\" width=\"100\"></iframe><p>", st)
	fmt.Fprint(w, "Set the recording time and buffer size, and select 'Start' to start the Flight Recorder")
	fmt.Fprint(w, "<form action=\""+flightRecorder+"/start\" target=\"response\">")
	fmt.Fprintf(w, "Recording time: <input type=\"number\" value=\"%d\" max=\"600\" min=\"1\" name=\"time\"> (seconds)<p>", recorderTime)
	fmt.Fprintf(w, "Buffer size: <input type=\"number\" value=\"%d\" max=\"100\" min=\"1\" name=\"buffer\"> (Mbytes)<p>", recorderSize)
	fmt.Fprintf(w, "<button type=\"submit\">Start recorder</button><br>")
	fmt.Fprintf(w, "</form>")
	fmt.Fprint(w, "<form action=\""+flightRecorder+"/stop\" target=\"response\">")
	fmt.Fprint(w, "<button type=\"submit\">Stop recorder</button><br>")
	fmt.Fprint(w, "</form>")
	fmt.Fprint(w, "<form action=\""+flightRecorder+"/download\" target=\"_blank\">")
	fmt.Fprint(w, "<button type=\"submit\">Download recorder data</button><br>")
	fmt.Fprint(w, "</form>")
	fmt.Fprint(w, "</body></html>")
}

func flightRecorderStart(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	recorderLock.Lock()
	defer recorderLock.Unlock()
	if recorder != nil {
		recorder.Stop()
	}
	recorderTime = parseNumber(r.Form["time"], 600, 10)
	recorderSize = parseNumber(r.Form["buffer"], 100, 5)
	recorder = trace.NewFlightRecorder(trace.FlightRecorderConfig{MinAge: time.Duration(recorderTime) * time.Second, MaxBytes: uint64(recorderSize) * 1024 * 1024})
	recorder.Start()
	fmt.Fprint(w, "RUNNING")
}

func flightRecorderStop(w http.ResponseWriter, r *http.Request) {
	recorderLock.Lock()
	defer recorderLock.Unlock()
	if recorder != nil {
		recorder.Stop()
		recorder = nil
	}
	fmt.Fprint(w, "STOPPED")
}

func flightRecorderDownload(w http.ResponseWriter, r *http.Request) {
	recorderLock.Lock()
	defer recorderLock.Unlock()
	if recorder != nil {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\"traces.out\"")
		recorder.WriteTo(w)
	} else {
		fmt.Fprintf(w, "Not recording\n")
	}
}

func parseNumber(s []string, max int, def int) int {
	if len(s) != 1 {
		return def
	}
	i, err := strconv.Atoi(s[0])
	if err != nil || i < 1 || i > max {
		return def
	}
	return i
}
