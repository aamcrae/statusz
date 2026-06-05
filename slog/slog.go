package slog

import (
	"log/slog"
	"sync"

	"github.com/aamcrae/statusz"
)

var replaceDefault sync.Once

func SlogDefault(held uint) {
	replaceDefault.Do(func() {
		slog.SetDefault(slog.New(slog.NewMultiHandler(slog.Default().Handler(), StatuszHandler(held))))
	})
}

func Slog(held uint) *slog.Logger {
	return slog.New(StatuszHandler(held))
}

func StatuszHandler(held uint) slog.Handler {
	return slog.NewTextHandler(statusz.MemLogger(held), nil)
}
