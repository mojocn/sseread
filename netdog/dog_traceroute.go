package netdog

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"
)

const (
	DEFAULT_PORT        = 33434
	DEFAULT_MAX_HOPS    = 64
	DEFAULT_FIRST_HOP   = 1
	DEFAULT_TIMEOUT_MS  = 500
	DEFAULT_RETRIES     = 3
	DEFAULT_PACKET_SIZE = 52
)

// Return the first non-loopback address as a 4 byte IP address. This address
// is used for sending packets out.
func socketAddr() (addr [4]byte, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if len(ipnet.IP.To4()) == net.IPv4len {
				copy(addr[:], ipnet.IP.To4())
				return
			}
		}
	}
	err = errors.New("no non-loopback IPv4 address found")
	return
}

// Given a host name convert it to a 4 byte IP address.
func destAddr(dest string) (addr [4]byte, err error) {
	addrs, err := net.LookupIP(dest)
	if err != nil {
		return
	}

	for _, ip := range addrs {
		if ipv4 := ip.To4(); ipv4 != nil {
			copy(addr[:], ipv4)
			return addr, nil
		}
	}
	return addr, fmt.Errorf("%q has no IPv4 address", dest)
}

// TracerouteOptions configures a traceroute.
type TracerouteOptions struct {
	port       int
	maxHops    int
	firstHop   int
	timeoutMs  int
	retries    int
	packetSize int
}

// TracrouteOptions is kept as an alias for compatibility with the original
// misspelled type name.
type TracrouteOptions = TracerouteOptions

func (options *TracerouteOptions) Port() int {
	if options.port == 0 {
		options.port = DEFAULT_PORT
	}
	return options.port
}

func (options *TracerouteOptions) SetPort(port int) {
	options.port = port
}

func (options *TracerouteOptions) MaxHops() int {
	if options.maxHops == 0 {
		options.maxHops = DEFAULT_MAX_HOPS
	}
	return options.maxHops
}

func (options *TracerouteOptions) SetMaxHops(maxHops int) {
	options.maxHops = maxHops
}

func (options *TracerouteOptions) FirstHop() int {
	if options.firstHop == 0 {
		options.firstHop = DEFAULT_FIRST_HOP
	}
	return options.firstHop
}

func (options *TracerouteOptions) SetFirstHop(firstHop int) {
	options.firstHop = firstHop
}

func (options *TracerouteOptions) TimeoutMs() int {
	if options.timeoutMs == 0 {
		options.timeoutMs = DEFAULT_TIMEOUT_MS
	}
	return options.timeoutMs
}

func (options *TracerouteOptions) SetTimeoutMs(timeoutMs int) {
	options.timeoutMs = timeoutMs
}

func (options *TracerouteOptions) Retries() int {
	if options.retries == 0 {
		options.retries = DEFAULT_RETRIES
	}
	return options.retries
}

func (options *TracerouteOptions) SetRetries(retries int) {
	options.retries = retries
}

func (options *TracerouteOptions) PacketSize() int {
	if options.packetSize == 0 {
		options.packetSize = DEFAULT_PACKET_SIZE
	}
	return options.packetSize
}

func (options *TracerouteOptions) SetPacketSize(packetSize int) {
	options.packetSize = packetSize
}

// TracerouteHop type
type TracerouteHop struct {
	Success     bool
	Address     [4]byte
	Host        string
	N           int
	ElapsedTime time.Duration
	TTL         int
}
type TraceRouteHopH struct {
	Success     bool          `json:"success"`
	Address     string        `json:"address"`
	Host        string        `json:"host"`
	N           int           `json:"n"`
	ElapsedTime time.Duration `json:"elapsed_time"`
	TTL         int           `json:"ttl"`
}

func (hop *TracerouteHop) AddressString() string {
	return fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
}

func (hop *TracerouteHop) HostOrAddressString() string {
	hostOrAddr := hop.AddressString()
	if hop.Host != "" {
		hostOrAddr = hop.Host
	}
	return hostOrAddr
}

// TracerouteResult type
type TracerouteResult struct {
	DestinationAddress [4]byte
	Hops               []TracerouteHop
}

func notify(hop TracerouteHop, channels []chan TracerouteHop) {
	for _, c := range channels {
		c <- hop
	}
}

func isReceiveTimeout(err error) bool {
	return errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK)
}

func traceAttempt(destination, local [4]byte, options *TracerouteOptions, ttl int) (TracerouteHop, bool, error) {
	receiveSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		return TracerouteHop{}, false, err
	}
	defer syscall.Close(receiveSocket)

	sendSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		return TracerouteHop{}, false, err
	}
	defer syscall.Close(sendSocket)

	if err := syscall.SetsockoptInt(sendSocket, syscall.IPPROTO_IP, syscall.IP_TTL, ttl); err != nil {
		return TracerouteHop{}, false, fmt.Errorf("set TTL: %w", err)
	}
	timeout := syscall.NsecToTimeval((time.Duration(options.TimeoutMs()) * time.Millisecond).Nanoseconds())
	if err := syscall.SetsockoptTimeval(receiveSocket, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &timeout); err != nil {
		return TracerouteHop{}, false, fmt.Errorf("set receive timeout: %w", err)
	}

	localAddress := &syscall.SockaddrInet4{Port: options.Port(), Addr: local}
	if err := syscall.Bind(receiveSocket, localAddress); err != nil {
		return TracerouteHop{}, false, fmt.Errorf("bind receive socket: %w", err)
	}

	destinationAddress := &syscall.SockaddrInet4{Port: options.Port(), Addr: destination}
	start := time.Now()
	if err := syscall.Sendto(sendSocket, []byte{0}, 0, destinationAddress); err != nil {
		return TracerouteHop{}, false, fmt.Errorf("send probe: %w", err)
	}

	packet := make([]byte, options.PacketSize())
	packetSize, from, err := syscall.Recvfrom(receiveSocket, packet, 0)
	if err != nil {
		if isReceiveTimeout(err) {
			return TracerouteHop{}, false, nil
		}
		return TracerouteHop{}, false, fmt.Errorf("receive response: %w", err)
	}
	remoteAddress, ok := from.(*syscall.SockaddrInet4)
	if !ok {
		return TracerouteHop{}, false, fmt.Errorf("unexpected address type %T", from)
	}

	hop := TracerouteHop{
		Success:     true,
		Address:     remoteAddress.Addr,
		N:           packetSize,
		ElapsedTime: time.Since(start),
		TTL:         ttl,
	}
	if hosts, lookupErr := net.LookupAddr(hop.AddressString()); lookupErr == nil && len(hosts) > 0 {
		hop.Host = hosts[0]
	}

	return hop, true, nil
}

func closeNotify(channels []chan TracerouteHop) {
	for _, c := range channels {
		close(c)
	}
}

// Traceroute uses the given dest (hostname) and options to execute a traceroute
// from your machine to the remote host.
//
// Outbound packets are UDP packets and inbound packets are ICMP.
//
// Returns a TracerouteResult which contains an array of hops. Each hop includes
// the elapsed time and its IP address.
func Traceroute(dest string, options *TracerouteOptions, c ...chan TracerouteHop) (result TracerouteResult, err error) {
	result.Hops = make([]TracerouteHop, 0)
	defer closeNotify(c)
	if options == nil {
		options = &TracerouteOptions{}
	}

	destinationAddress, err := destAddr(dest)
	if err != nil {
		return result, err
	}
	result.DestinationAddress = destinationAddress
	localAddress, err := socketAddr()
	if err != nil {
		return result, err
	}

	ttl := options.FirstHop()
	retry := 0
	for {
		hop, received, err := traceAttempt(destinationAddress, localAddress, options, ttl)
		if err != nil {
			return result, err
		}
		if received {
			notify(hop, c)
			result.Hops = append(result.Hops, hop)
			retry = 0

			if ttl >= options.MaxHops() || hop.Address == destinationAddress {
				return result, nil
			}
			ttl++
		} else {
			retry += 1
			if retry > options.Retries() {
				notify(TracerouteHop{Success: false, TTL: ttl}, c)
				ttl += 1
				retry = 0
			}

			if ttl >= options.MaxHops() {
				return result, nil
			}
		}

	}
}

type TraceRouteResultH struct {
	Hops               []TraceRouteHopH `json:"hops,omitempty"`
	DestinationAddress string           `json:"destination_address,omitempty"`
	Error              error            `json:"error,omitempty"`
}

func TraceRouteRun(dest string) TraceRouteResultH {
	options := &TracerouteOptions{}
	options.SetMaxHops(16)
	options.SetRetries(2)
	options.SetTimeoutMs(500)

	hops := make(chan TracerouteHop, options.MaxHops())
	result, err := Traceroute(dest, options, hops)
	if err != nil {
		return TraceRouteResultH{
			Error: err,
		}
	}
	return TraceRouteResultH{
		Hops:               convertHops(result.Hops),
		DestinationAddress: fmt.Sprintf("%v.%v.%v.%v", result.DestinationAddress[0], result.DestinationAddress[1], result.DestinationAddress[2], result.DestinationAddress[3]),
	}
}
func convertHops(hops []TracerouteHop) []TraceRouteHopH {
	result := make([]TraceRouteHopH, len(hops))
	for i, hop := range hops {
		result[i] = TraceRouteHopH{
			Success:     hop.Success,
			Address:     hop.AddressString(),
			Host:        hop.Host,
			N:           hop.N,
			ElapsedTime: hop.ElapsedTime,
			TTL:         hop.TTL,
		}
	}
	return result
}
