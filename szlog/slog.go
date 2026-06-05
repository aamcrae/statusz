// Package slog provides a log/slog handler that saves logs
// into a circular memory buffer.
package szlog

import (
	"log/slog"

	"github.com/aamcrae/statusz"
)

// Slog creates a structure logger with a memory buffer handler
func Slog(held uint) *slog.Logger {
	return slog.New(StatuszHandler(held))
}

// StatuszHandler creates a Handler for the memory logger
func StatuszHandler(held uint) slog.Handler {
	return slog.NewTextHandler(statusz.MemLogger(held), nil)
}
