package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/mojocn/sseread/netdog"
)

// standard go http.Handler
func myStdHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /netdog-http", netdog.HandlerDogHTTP)
	mux.HandleFunc("POST /netdog-network", netdog.HandlerDogNetwork)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

// netlify lambda entry point, wrap std http.Handler
func main() {
	adapter := httpadapter.New(myStdHandler())
	lambda.Start(adapter.Proxy)
}
