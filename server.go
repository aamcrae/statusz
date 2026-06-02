package statusz

import (
	"net/http"
)

// Serve creates a HTTP server on the address passed, and installs
// the statusz handler on it. This only returns when the server stops.
func Serve(addr string) error {
	mux := http.NewServeMux()
	Install(mux)
	return http.ListenAndServe(addr, mux)
}
