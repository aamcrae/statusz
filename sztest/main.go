package main

import (
	"flag"
	"fmt"
	"net/http"
)

var port = flag.Int("port", 9000, "Server port")

func main() {

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		panic(err)
	}
}
