package szlog

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aamcrae/statusz"
)

func TestSlog(t *testing.T) {
	logger := Slog(5)
	for i := range 5 {
		logger.Info("slog number", "val", i)
	}
	m := http.NewServeMux()
	statusz.Install(m)
	req, err := http.NewRequest("GET", "/statusz", nil)
	if err != nil {
		t.Fatal(err)
	}
	h, _ := m.Handler(req)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	str := recorder.Body.String()
	for i := range 5 {
		exp := fmt.Sprintf("msg=\"slog number\" val=%d", i)
		if strings.Index(str, exp) == -1 {
			t.Errorf("Missing log line: got %s, wanted %s", str, exp)
		}
	}
}
