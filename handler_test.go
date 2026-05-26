//go:build unit

package statusz

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const localLine = "Local handler line<br>"

func TestStatuszHandler(t *testing.T) {
	RegisterExtension(localHandlerTest)
	RegisterPage("My page", "mypage", func(http.ResponseWriter, *http.Request) {})
	req, err := http.NewRequest("GET", "/statusz", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	handler := http.HandlerFunc(statuszHandler)
	handler.ServeHTTP(resp, req)
	status := resp.Code
	if status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}
	str := resp.Body.String()
	ind := strings.Index(str, localLine)
	if ind == -1 {
		t.Errorf("Local handler string not found")
	}
	ind = strings.Index(str, "<a href=\"/statusz/mypage\">My page</a>")
	if ind == -1 {
		t.Errorf("My page link not found")
	}
}

func TestInstall(t *testing.T) {
	RegisterExtension(localHandlerTest)
	hnd := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "muxpage")
	}
	RegisterPage("Mux page", "muxpage", hnd)
	m := http.NewServeMux()
	Install(m)
	req, err := http.NewRequest("GET", "/statusz/muxpage", nil)
	if err != nil {
		t.Fatal(err)
	}
	h, pattern := m.Handler(req)
	if pattern != "/statusz/muxpage" {
		t.Errorf("Mux returned wrong pattern : got %v, want %v", pattern, "/statusz/muxpage")
	}
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	str := recorder.Body.String()
	if str != "muxpage" {
		t.Errorf("Page returned wrong data : got %v, want %v", str, "muxpage")
	}
}

func localHandlerTest(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, localLine)
}
