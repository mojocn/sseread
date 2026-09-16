package netdog

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func checkHttp(method, url string, headers map[string]string, body []byte, timeout time.Duration) (cost time.Duration, response *http.Response, err error) {
	url = strings.TrimSpace(url)
	start := time.Now()
	client := *httpClient
	client.Timeout = timeout
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 26_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/26.0 Mobile/15E148 Safari/604.1")
	resp, err := client.Do(req)
	return time.Since(start), resp, err

}
