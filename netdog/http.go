package netdog

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func returnResponse(w http.ResponseWriter, result any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

func HandlerDogNetwork(w http.ResponseWriter, r *http.Request) {
	body := new(DogWatchRequestNetwork)
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		result := DogWatchResult{
			Error: fmt.Errorf("failed to decode JSON body: %v", err),
		}
		returnResponse(w, result)
		return
	}
	result := DogWatchNetwork(body)
	returnResponse(w, result)
}

func HandlerDogHTTP(w http.ResponseWriter, r *http.Request) {
	body := new(DogWatchRequestHTTP)
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		returnResponse(w, DogWatchResult{
			Error: fmt.Errorf("failed to decode JSON body: %v", err),
		})
		return
	}
	result := DogWatchHttp(body)
	returnResponse(w, result)
}
