package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aamcrae/statusz/szlog"
)

var port = flag.Int("port", 9000, "Server port")

func main() {

	// Default logging to stdout and statusz
	sl := slog.New(slog.NewMultiHandler(szlog.StatuszHandler(10), slog.NewTextHandler(os.Stdout, nil)))
	slog.SetDefault(sl)
	
	go func() {
		var i int
		for {
			i++
			slog.Info("slog number", "val", i)
			time.Sleep(1 * time.Second)
		}
	}()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		panic(err)
	}
}
