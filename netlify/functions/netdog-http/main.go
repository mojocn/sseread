package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mojocn/sseread/netdog"
)

const timeout = 45 * time.Second

func responseJSON(result any) *events.APIGatewayProxyResponse {
	w := &events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type": "application/json; charset=utf-8",
		},
		StatusCode: http.StatusOK,
		Body:       "",
	}
	if resultBytes, err := json.Marshal(result); err == nil {
		w.Body = string(resultBytes)
	}
	return w
}

func handler(r events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	body := new(netdog.DogWatchRequestHTTP)
	if err := json.Unmarshal([]byte(r.Body), body); err != nil {
		return responseJSON(map[string]string{"error": err.Error()}), nil
	}
	result := netdog.DogWatchHttp(body)
	return responseJSON(result), nil
}

func main() {
	// Make the handler available for Remote Procedure Call
	lambda.Start(handler)
}
