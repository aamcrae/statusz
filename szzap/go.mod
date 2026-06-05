module github.com/aamcrae/statusz/szzap

go 1.25.0

replace github.com/aamcrae/statusz => ../../statusz

require (
	github.com/aamcrae/statusz v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.28.0
)

require go.uber.org/multierr v1.10.0 // indirect
