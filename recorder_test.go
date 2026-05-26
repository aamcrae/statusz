//go:build unit

package statusz

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecorder(t *testing.T) {
	start, err := http.NewRequest("GET", "/start?time=20&buffer=3", nil)
	if err != nil {
		t.Fatal(err)
	}
	startResp := httptest.NewRecorder()
	startHandler := http.HandlerFunc(flightRecorderStart)
	startHandler.ServeHTTP(startResp, start)
	if startResp.Code != http.StatusOK {
		t.Errorf("Recorder start returned wrong status code: got %v, want %v", startResp.Code, http.StatusOK)
	}
	if recorder == nil {
		t.Fatalf("Recorder not started")
	}
	if !recorder.Enabled() {
		t.Errorf("Recorder present but not enabled")
	}
	if recorderTime != 20 {
		t.Errorf("Recorder start time wrong: got %v, want %v", recorderTime, 20)
	}
	if recorderSize != 3 {
		t.Errorf("Recorder buffer size wrong: got %v, want %v", recorderSize, 3)
	}
	dl, err := http.NewRequest("GET", "/download", nil)
	if err != nil {
		t.Fatal(err)
	}
	dlResp := httptest.NewRecorder()
	dlHandler := http.HandlerFunc(flightRecorderDownload)
	dlHandler.ServeHTTP(dlResp, dl)
	if dlResp.Code != http.StatusOK {
		t.Errorf("Recorder download returned wrong status code: got %v, want %v", dlResp.Code, http.StatusOK)
	}
	// Recorder still running
	if recorder == nil || !recorder.Enabled() {
		t.Errorf("Recorder stopped after download")
	}
	ct := dlResp.Result().Header.Get("Content-Type")
	if ct != "application/octet-stream" {
		t.Errorf("Recorder download wrong content type : got %v, want %v", ct, "application/octet-stream")
	}
	stop, err := http.NewRequest("GET", "/stop", nil)
	if err != nil {
		t.Fatal(err)
	}
	stopResp := httptest.NewRecorder()
	stopHandler := http.HandlerFunc(flightRecorderStop)
	stopHandler.ServeHTTP(stopResp, stop)
	if stopResp.Code != http.StatusOK {
		t.Errorf("Recorder stop returned wrong status code: got %v, want %v", stopResp.Code, http.StatusOK)
	}
	if recorder != nil {
		t.Errorf("Recorder not stopped")
	}
}
