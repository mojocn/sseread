package main

import (
	"encoding/json"
	"fmt"
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
	dest := r.QueryStringParameters["dest"]
	if dest == "" {
		return responseJSON(netdog.TraceRouteResultH{
			Error: fmt.Errorf("missing 'dest' query parameter"),
		}), nil
	}
	result := netdog.TraceRouteRun(dest)
	return responseJSON(result), nil
}

func main() {
	// Make the handler available for Remote Procedure Call
	lambda.Start(handler)
}
