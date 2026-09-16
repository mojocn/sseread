package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mojocn/sseread/netdog"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /netdog-http", netdog.HandlerDogHTTP)
	mux.HandleFunc("POST /netdog-network", netdog.HandlerDogNetwork)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
