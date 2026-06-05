package szzap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aamcrae/statusz"
	"go.uber.org/zap"
)

func TestZap(t *testing.T) {
	err := ZapRegister(5)
	if err != nil {
		t.Fatal(err)
	}
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{StatuszSink + "://"}
	logger, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}
	for i := range 5 {
		logger.Info("zap number", zap.Int("val", i))
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
		exp := fmt.Sprintf("\"msg\":\"zap number\",\"val\":%d", i)
		if strings.Index(str, exp) == -1 {
			t.Errorf("Missing log line: got %s, wanted %s", str, exp)
		}
	}
}
