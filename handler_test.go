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

	RegisterLocalHandler(localHandlerTest)
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
	ind := strings.Index(resp.Body.String(), localLine)
	if ind == -1 {
		t.Errorf("Local handler string not found")
	}
}

func localHandlerTest(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, localLine)
}
