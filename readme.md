[![GoDoc](https://pkg.go.dev/badge/github.com/mojocn/sseread?status.svg)](https://pkg.go.dev/github.com/mojocn/sseread?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/mojocn/sseread?)](https://goreportcard.com/report/github.com/mojocn/sseread)
[![codecov](https://codecov.io/gh/mojocn/sseread/graph/badge.svg?token=3UC1L5K4LY)](https://codecov.io/gh/mojocn/sseread)
[![Go version](https://img.shields.io/github/go-mod/go-version/mojocn/sseread.svg)](https://github.com/mojocn/sseread)
[![Follow mojocn](https://img.shields.io/github/followers/mojocn?label=Follow&style=social)](https://github.com/mojocn)


# Server Sent Events Reader

This is a simple library of how to read Server Sent Events (SSE) stream from `Response.Body` in Golang.


## Usage
download the library using
`go get -u github.com/mojocn/sseread@latest`

simple examples of how to use the library.

1. [read SSE by callback](https://pkg.go.dev/github.com/mojocn/sseread#example-Read) 
2. [read SSE by channel](https://pkg.go.dev/github.com/mojocn/sseread#example-ReadCh)
3. [cloudflare AI text generation example](cloudflare_ai_test.go)

```go

## Testing

```bash
# git clone https://github.com/mojocn/sseread.git && cd sseread
go test -v
```





