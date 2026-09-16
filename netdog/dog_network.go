package netdog

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"
)

func checkNetwork(tcpOrUdp, hostnameOrIP, port string, timeout time.Duration, checkTLS bool) (cost time.Duration, cert *x509.Certificate, err error) {
	tcpOrUdp = strings.ToLower(tcpOrUdp)
	if tcpOrUdp != "tcp" && tcpOrUdp != "udp" {
		return 0, nil, fmt.Errorf("unsupported network type: %s", tcpOrUdp)
	}
	if tcpOrUdp == "udp" && checkTLS {
		checkTLS = false
	}
	hostPort := net.JoinHostPort(hostnameOrIP, port)
	start := time.Now()
	conn, err := net.DialTimeout(tcpOrUdp, hostPort, timeout)
	if err != nil {
		return 0, nil, err
	}
	defer conn.Close()
	if checkTLS {
		tlsConn := tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true, // 仅抓取证书，不做证书链校验
			ServerName:         hostnameOrIP,
		})
		defer tlsConn.Close()

		if err := tlsConn.SetDeadline(time.Now().Add(timeout)); err != nil {
			return 0, nil, err
		}
		if err := tlsConn.Handshake(); err != nil {
			return 0, nil, err
		}
		state := tlsConn.ConnectionState()
		if len(state.PeerCertificates) > 0 {
			return time.Since(start), state.PeerCertificates[0], nil
		}
		return time.Since(start), nil, fmt.Errorf("no peer certificate found")
	}
	return time.Since(start), cert, nil
}
