# go-binenc-gen

## Overview

GoDoc @ https://pkg.go.dev/github.com/cezarguimaraes/go-binenc-gen

## TODOs

- Fix `go vet` complaints about `WriteTo` and `ReadFrom` signature
- Add error handling to `ReadFrom` and count bytes read
- Add benchmark for integer and float heavy datasets
    - Possibly do the same allocation optimization (for slices/arrays of these) done for strings
- Optimize `[]byte` read/writes
- Add endianness option
- Add type filter option
- Add output option to suport multiple `go generate` on the same package

## Latest benchmarks

```
go test -bench=BenchmarkConda* -v -benchmem
=== RUN   TestCondaRead
--- PASS: TestCondaRead (1.20s)
goos: linux
goarch: amd64
pkg: github.com/cezarguimaraes/go-binenc-gen/bench
cpu: 12th Gen Intel(R) Core(TM) i7-12700KF
BenchmarkCondaBinencRead
BenchmarkCondaBinencRead-20                  176           6692881 ns/op         998.28 MB/s    14835038 B/op      26668 allocs/op
BenchmarkCondaJSONRead
BenchmarkCondaJSONRead-20                     12          91524839 ns/op         115.43 MB/s    63849491 B/op     521843 allocs/op
BenchmarkCondaGobRead
BenchmarkCondaGobRead-20                      79          14271872 ns/op         456.87 MB/s    20436163 B/op     393282 allocs/op
BenchmarkCondaProtobufRead
BenchmarkCondaProtobufRead-20                 79          14893087 ns/op         442.22 MB/s    17447636 B/op     415400 allocs/op
BenchmarkCondaBinencWrite
    conda_bench_test.go:192: binenc write size: 6681393
    conda_bench_test.go:192: binenc write size: 6681393
    conda_bench_test.go:192: binenc write size: 6681393
BenchmarkCondaBinencWrite-20                 276           4240484 ns/op        1575.62 MB/s     6708907 B/op          1 allocs/op
BenchmarkCondaJSONWrite
    conda_bench_test.go:206: json write size: 10564362
    conda_bench_test.go:206: json write size: 10564362
    conda_bench_test.go:206: json write size: 10564362
BenchmarkCondaJSONWrite-20                    74          15860022 ns/op         666.10 MB/s      596432 B/op          1 allocs/op
BenchmarkCondaGobWrite
    conda_bench_test.go:220: gob write size: 6520461
    conda_bench_test.go:220: gob write size: 6520461
    conda_bench_test.go:220: gob write size: 6520461
BenchmarkCondaGobWrite-20                    103          11338828 ns/op         575.06 MB/s    33962477 B/op      26741 allocs/op
BenchmarkCondaGobWriteLenient
    conda_bench_test.go:234: gob write size: 6520461
    conda_bench_test.go:234: gob write size: 6519990
    conda_bench_test.go:234: gob write size: 6519990
BenchmarkCondaGobWriteLenient-20             172           6930366 ns/op         940.79 MB/s      870913 B/op      26647 allocs/op
BenchmarkCondaProtobuWrite
    conda_bench_test.go:249: protobuf write size: 6586030
    conda_bench_test.go:249: protobuf write size: 6586030
    conda_bench_test.go:249: protobuf write size: 6586030
BenchmarkCondaProtobuWrite-20                152           7755885 ns/op         849.17 MB/s     6586396 B/op          1 allocs/op
PASS
ok      github.com/cezarguimaraes/go-binenc-gen/bench   17.449s
```
