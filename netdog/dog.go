// Package netdog provides HTTP and network connectivity checks.
package netdog

import (
	"crypto/x509"
	"errors"
	"io"
	"net/http"
	"time"
)

// DogWatchRequestHTTP describes an HTTP request to check.
type DogWatchRequestHTTP struct {
	// Method is the HTTP method, such as GET or POST.
	Method string `json:"method"`
	// URL is the endpoint to request.
	URL string `json:"url"`
	// Headers contains the request headers.
	Headers map[string]string `json:"headers"`
	// Body contains the request body.
	Body []byte `json:"body"`
	// Timeout is clamped to the range from one second to one minute.
	Timeout time.Duration `json:"timeout"` //1~60s
}

// DogWatchResultHTTP contains the result of an HTTP connectivity check.
type DogWatchResultHTTP struct {
	DogWatchResult
	// Headers contains the response headers.
	Headers http.Header `json:"headers"`
	// Body contains up to 2 MiB of the response body.
	Body []byte `json:"body"`
}

const maxHTTPResponseBodyBytes = 2 << 20

const (
	minTimeout = time.Second
	maxTimeout = 60 * time.Second
)

func clampTimeout(timeout time.Duration) time.Duration {
	if timeout < minTimeout {
		return minTimeout
	}
	if timeout > maxTimeout {
		return maxTimeout
	}
	return timeout
}

// DogWatchHttp performs an HTTP connectivity check and returns its response metadata and body.
func DogWatchHttp(param *DogWatchRequestHTTP) DogWatchResultHTTP {
	if param == nil {
		return DogWatchResultHTTP{DogWatchResult: DogWatchResult{Error: errors.New("HTTP request is nil")}}
	}
	timeout := clampTimeout(param.Timeout)

	cost, response, err := checkHttp(param.Method, param.URL, param.Headers, param.Body, timeout)
	result := DogWatchResultHTTP{
		DogWatchResult: DogWatchResult{
			Cost:  cost,
			Error: err,
		},
	}
	if response != nil {
		result.Headers = response.Header
		defer response.Body.Close()
		result.Body, err = io.ReadAll(io.LimitReader(response.Body, maxHTTPResponseBodyBytes))
		if result.Error == nil && err != nil {
			result.Error = err
		}
	}
	if response != nil && response.TLS != nil && len(response.TLS.PeerCertificates) > 0 {
		setCertificateMetadata(&result.DogWatchResult, response.TLS.PeerCertificates[0])
	}
	return result
}

// DogWatchRequestNetwork describes a TCP or UDP connectivity check.
type DogWatchRequestNetwork struct {
	// Network is the network type: tcp or udp.
	Network string `json:"network"` // tcp or udp
	// Host is a hostname or IP address.
	Host string `json:"host"` //host or ip
	// Port is the destination port number.
	Port string `json:"port"` //port number eg 80 443
	// Timeout is clamped to the range from one second to one minute.
	Timeout time.Duration `json:"timeout"` //1~60s
	// TLS controls whether a TLS handshake is performed after connecting.
	TLS bool `json:"tls"` // whether to check TLS
}

// DogWatchResult contains the result of a network connectivity check.
type DogWatchResult struct {
	// Cost is the time spent performing the check.
	Cost time.Duration `json:"cost"`
	// TlsIssuer is the issuer of the peer certificate, when TLS is checked.
	TlsIssuer string `json:"tls_issuer"`
	// TlsSubject is the subject of the peer certificate, when TLS is checked.
	TlsSubject string `json:"tls_subject"`
	// TlsNotBefore is the start of the peer certificate validity period.
	TlsNotBefore time.Time `json:"tls_not_before"`
	// TlsNotAfter is the end of the peer certificate validity period.
	TlsNotAfter time.Time `json:"tls_not_after"`
	// Error is the error encountered while performing the check, if any.
	Error error `json:"error"`
}

// DogWatchNetwork performs a TCP or UDP connectivity check and returns its result.
func DogWatchNetwork(param *DogWatchRequestNetwork) DogWatchResult {
	if param == nil {
		return DogWatchResult{Error: errors.New("network request is nil")}
	}
	timeout := clampTimeout(param.Timeout)

	cost, cert, err := checkNetwork(param.Network, param.Host, param.Port, timeout, param.TLS)
	result := DogWatchResult{
		Cost:  cost,
		Error: err,
	}
	setCertificateMetadata(&result, cert)
	return result
}

func setCertificateMetadata(result *DogWatchResult, cert *x509.Certificate) {
	if cert == nil {
		return
	}
	result.TlsIssuer = cert.Issuer.String()
	result.TlsSubject = cert.Subject.String()
	result.TlsNotBefore = cert.NotBefore
	result.TlsNotAfter = cert.NotAfter
}
