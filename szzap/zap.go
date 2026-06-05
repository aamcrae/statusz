// Package szzap provides a Zap logging sink for the statusz
// memory logger. This is a separate package to avoid
// dependency on zap for the main portion of statusz.
package szzap

import (
	"io"
	"net/url"

	"github.com/aamcrae/statusz"
	"go.uber.org/zap"
)

const StatuszSink = "statusz"

type szZap struct {
	io.Writer
}

func (z *szZap) Close() error {
	return nil
}

func (z *szZap) Sync() error {
	return nil
}

// ZapRegister creates a sink to send logs to the statusz memory logger.
func ZapRegister(held uint) error {
	return zap.RegisterSink(StatuszSink, func(*url.URL) (zap.Sink, error) {
		return &szZap{statusz.MemLogger(held)}, nil
	})
}
