# Introduction

A `Koa` implementation in Go Language. Inspired by its middleware mechanism, I designed overall framework purely by Go.


# Environment
OS: Win11 x64
Go: 1.21+

# Build Example

I'm lazy, how to use please see the sample code.

Take `example/serve` for example:
```shell
go build -o build/serve.exe github.com/clouddea/koa-go/example/serve
./build/serve --help
./build/serve
```

# TODOs
+ [x] finish implementation of `Context` API
+ [ ] finish implementation of `Config` API
+ [ ] finish implementation of `Request` API
+ [ ] finish implementation of `Response` API
+ [ ] finish implementation of other APIs

# Reference

## benchmark
77492.47 request/s in writing "hello world"
```shell
go install github.com/tsliwowicz/go-wrk@latest
go-wrk -help
go-wrk -c 128 -d 10 http://localhost:8080/test  # amd 7840h laptop
```







