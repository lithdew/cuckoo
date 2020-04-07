# cuckoo

[![MIT License](https://img.shields.io/apm/l/atomic-design-ui.svg?)](LICENSE)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/lithdew/cuckoo)
[![Discord Chat](https://img.shields.io/discord/697002823123992617)](https://discord.gg/58dJzS)

A fast, vectorized Cuckoo filter implementation in Go with zero allocations in hot paths and in encoding to/decoding from its in-memory representation.

Out-of-the-box, cuckoo comes with nice defaults for the purposes of state reconciliation across a distributed system. Though, feel free to configure the total number of buckets/number of bytes per bucket in [filter.go](filter.go) to taste for your specific application.

Refer to [filter_test.go](filter_test.go) for usage instructions.

## Benchmarks

```
go test -bench=. -benchtime=10s -benchmem

goos: linux
goarch: amd64
pkg: github.com/lithdew/cuckoo
BenchmarkNewFilter-8                       39187            290737 ns/op         2105346 B/op          1 allocs/op
BenchmarkInsert-8                       239464290               51.8 ns/op             0 B/op          0 allocs/op
BenchmarkLookup-8                       238444526               49.7 ns/op             0 B/op          0 allocs/op
BenchmarkMarshalBinary-8                   24799            468098 ns/op         2097161 B/op          1 allocs/op
BenchmarkUnsafeUnmarshalBinary-8            6715           2097175 ns/op         2105777 B/op          5 allocs/op
BenchmarkUnmarshalBinary-8                  5738           2042495 ns/op         2105777 B/op          5 allocs/op
```