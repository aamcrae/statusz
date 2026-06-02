package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aamcrae/statusz"
)

var port = flag.Int("port", 9000, "Server port")

func main() {

	statusz.Logs(10)
	go func() {
		var i int
		for {
			i++
			log.Printf("log number %d", i)
			time.Sleep(2 * time.Second)
		}
	}()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		panic(err)
	}
}
