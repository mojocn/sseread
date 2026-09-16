package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
)

func handler(r events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Your server-side functionality
	//get client ip
	//ip := r.RequestContext.Identity.SourceIP
	result := events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "application/json"},
		IsBase64Encoded: false,
	}

	rr := map[string]any{
		"x-forwarded-for":           r.Headers["x-forwarded-for"],
		"x-nf-client-connection-ip": r.Headers["x-nf-client-connection-ip"],
		"x-country":                 r.Headers["x-country"],
		"x-language":                r.Headers["x-language"],
		"user-agent":                r.Headers["user-agent"],
		"sec-ch-ua-platform":        r.Headers["sec-ch-ua-platform"],
		"sec-ch-ua-mobile":          r.Headers["sec-ch-ua-mobile"],
	}
	geoS := r.Headers["x-nf-geo"]
	if geoS != "" {
		//decode geoS base64
		geo, err := base64.StdEncoding.DecodeString(geoS)
		if err != nil {
			log.Println("decode geoS error", err)
		}
		geoInfo := map[string]any{}
		err = json.Unmarshal(geo, &geoInfo)
		if err != nil {
			log.Println("unmarshal geo error", err)
		} else {
			rr["geo"] = geoInfo
		}
	}
	marshal, err := json.MarshalIndent(rr, "", "      ")
	if err != nil {
		return nil, err
	}
	result.Body = string(marshal)
	return &result, nil
}

func main() {
	// Make the handler available for Remote Procedure Call
	lambda.Start(handler)
}
