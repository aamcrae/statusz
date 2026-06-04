package statusz

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildinfo(t *testing.T) {
	req, err := http.NewRequest("GET", "/buildinfo", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	biHandler := http.HandlerFunc(buildInfoHandler)
	biHandler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Build Info Handler returned wrong status code: got %v, want %v", resp.Code, http.StatusOK)
	}
	ind := strings.Index(resp.Body.String(), "Go toolchain version:")
	if ind == -1 {
		t.Errorf("Build information not retrieved")
	}
}
