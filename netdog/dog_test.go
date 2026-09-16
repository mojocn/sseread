package netdog

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDogWatchHttp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if got := r.Header.Get("X-Test"); got != "header-value" {
			t.Errorf("X-Test header = %q, want header-value", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		if string(body) != "request-body" {
			t.Errorf("request body = %q, want request-body", body)
		}
		w.Header().Set("X-Response", "response-value")
		_, _ = w.Write([]byte("response-body"))
	}))
	defer server.Close()

	request := &DogWatchRequestHTTP{
		Method:  http.MethodPost,
		URL:     server.URL,
		Headers: map[string]string{"X-Test": "header-value"},
		Body:    []byte("request-body"),
		Timeout: 100 * time.Millisecond,
	}
	result := DogWatchHttp(request)

	if result.Error != nil {
		t.Fatalf("DogWatchHttp returned error: %v", result.Error)
	}
	if string(result.Body) != "response-body" {
		t.Errorf("response body = %q, want response-body", result.Body)
	}
	if result.Headers.Get("X-Response") != "response-value" {
		t.Errorf("X-Response header = %q, want response-value", result.Headers.Get("X-Response"))
	}
	if request.Timeout != 100*time.Millisecond {
		t.Errorf("timeout = %s, want input unchanged", request.Timeout)
	}
	if result.Cost <= 0 {
		t.Errorf("cost = %s, want positive duration", result.Cost)
	}
}

func TestDogWatchHttpClampsTimeoutAndLimitsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", maxHTTPResponseBodyBytes+1)))
	}))
	defer server.Close()

	request := &DogWatchRequestHTTP{URL: server.URL, Timeout: 61 * time.Second}
	result := DogWatchHttp(request)

	if result.Error != nil {
		t.Fatalf("DogWatchHttp returned error: %v", result.Error)
	}
	if request.Timeout != 61*time.Second {
		t.Errorf("timeout = %s, want input unchanged", request.Timeout)
	}
	if len(result.Body) != maxHTTPResponseBodyBytes {
		t.Errorf("response body length = %d, want %d", len(result.Body), maxHTTPResponseBodyBytes)
	}
}

func TestDogWatchHttpInvalidURL(t *testing.T) {
	request := &DogWatchRequestHTTP{Method: http.MethodGet, URL: "://invalid", Timeout: time.Second}
	result := DogWatchHttp(request)

	if result.Error == nil {
		t.Fatal("DogWatchHttp returned nil error for invalid URL")
	}
	if result.Body != nil {
		t.Errorf("response body = %q, want nil", result.Body)
	}
}

func TestDogWatchNetworkTCP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	address := listener.Addr().String()
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatalf("split listener address: %v", err)
	}
	request := &DogWatchRequestNetwork{
		Network: "TCP",
		Host:    host,
		Port:    port,
		Timeout: 61 * time.Second,
	}
	result := DogWatchNetwork(request)

	if result.Error != nil {
		t.Fatalf("DogWatchNetwork returned error: %v", result.Error)
	}
	if request.Timeout != 61*time.Second {
		t.Errorf("timeout = %s, want input unchanged", request.Timeout)
	}
	if result.Cost <= 0 {
		t.Errorf("cost = %s, want positive duration", result.Cost)
	}
}

func TestDogWatchNetworkTLS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	address := strings.TrimPrefix(server.URL, "https://")
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatalf("split TLS server address: %v", err)
	}
	result := DogWatchNetwork(&DogWatchRequestNetwork{
		Network: "tcp",
		Host:    host,
		Port:    port,
		Timeout: time.Second,
		TLS:     true,
	})

	if result.Error != nil {
		t.Fatalf("DogWatchNetwork returned error: %v", result.Error)
	}
	if result.TlsSubject == "" || result.TlsIssuer == "" {
		t.Errorf("TLS certificate metadata is empty: issuer=%q subject=%q", result.TlsIssuer, result.TlsSubject)
	}
}

func TestCheckNetworkRejectsUnsupportedNetwork(t *testing.T) {
	_, cert, err := checkNetwork("icmp", "127.0.0.1", "0", time.Second, false)

	if err == nil {
		t.Fatal("checkNetwork returned nil error for unsupported network")
	}
	if !strings.Contains(err.Error(), "unsupported network type") {
		t.Errorf("error = %q, want unsupported network type", err)
	}
	if cert != nil {
		t.Errorf("certificate = %v, want nil", cert)
	}
}

func TestCheckNetworkUDPDisablesTLS(t *testing.T) {
	connection, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("listen UDP: %v", err)
	}
	defer connection.Close()

	resultCost, cert, err := checkNetwork("udp", "127.0.0.1", fmt.Sprint(connection.LocalAddr().(*net.UDPAddr).Port), time.Second, true)
	if err != nil {
		t.Fatalf("checkNetwork returned error: %v", err)
	}
	if resultCost <= 0 {
		t.Errorf("cost = %s, want positive duration", resultCost)
	}
	if cert != nil {
		t.Errorf("certificate = %v, want nil", cert)
	}
}

func TestCheckHttpSetsUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() == "" {
			t.Error("User-Agent header is empty")
		}
	}))
	defer server.Close()

	_, response, err := checkHttp(http.MethodGet, server.URL, nil, nil, time.Second)
	if err != nil {
		t.Fatalf("checkHttp returned error: %v", err)
	}
	if response == nil {
		t.Fatal("checkHttp returned nil response")
	}
	response.Body.Close()
}
