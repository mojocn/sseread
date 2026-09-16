package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mojocn/sseread/netdog"
)

// for FAAS of Vercel
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /netdog-http", netdog.HandlerDogHTTP)
	mux.HandleFunc("POST /netdog-network", netdog.HandlerDogNetwork)
	mux.HandleFunc("GET /ip", func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = ip
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(clientIP))
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
