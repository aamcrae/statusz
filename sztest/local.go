package main

import (
	"fmt"
	"net/http"

	"github.com/aamcrae/statusz"
)

func init() {
	statusz.RegisterLocalHandler(local)
	statusz.RegisterPage("pr/", prPage)
}

func local(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h2>Local statusz extension<h2>")
}

func prPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Extension page: %s", r.URL.Path)
}
