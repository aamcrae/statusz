package slog

import (
	"log/slog"

	"github.com/aamcrae/statusz"
)

func Slog(held uint) *slog.Logger {
	return slog.New(StatuszHandler(held))
}

func StatuszHandler(held uint) slog.Handler {
	return slog.NewTextHandler(statusz.MemLogger(held), nil)
}
