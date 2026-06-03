package statusz

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLog(t *testing.T) {
	logger := log.New(io.Discard, "test", log.LstdFlags)
	StdLogger(logger, 5)
	for i := range 5 {
		logger.Printf("number %d", i)
	}
	req, err := http.NewRequest("GET", "/statusz", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	logsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memLog.logsExtension(w)
	})
	logsHandler.ServeHTTP(resp, req)
	body := resp.Body.String()
	for i := range 5 {
		exp := fmt.Sprintf("number %d", i)
		if strings.Index(resp.Body.String(), exp) == -1 {
			t.Errorf("Missing log line: got %s, wanted %s", body, exp)
		}
	}
	logger.Print("number 5")
	req, err = http.NewRequest("GET", "/statusz", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp = httptest.NewRecorder()
	logsHandler.ServeHTTP(resp, req)
	body = resp.Body.String()
	if strings.Index(resp.Body.String(), "number 0") != -1 {
		t.Errorf("log line 0 not removed: got %s", body)
	}
}
