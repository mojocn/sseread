package netdog

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestTracerouteOptionsDefaultsAndSetters(t *testing.T) {
	options := &TracerouteOptions{}

	if got := options.Port(); got != DEFAULT_PORT {
		t.Errorf("Port() = %d, want %d", got, DEFAULT_PORT)
	}
	if got := options.MaxHops(); got != DEFAULT_MAX_HOPS {
		t.Errorf("MaxHops() = %d, want %d", got, DEFAULT_MAX_HOPS)
	}
	if got := options.FirstHop(); got != DEFAULT_FIRST_HOP {
		t.Errorf("FirstHop() = %d, want %d", got, DEFAULT_FIRST_HOP)
	}
	if got := options.TimeoutMs(); got != DEFAULT_TIMEOUT_MS {
		t.Errorf("TimeoutMs() = %d, want %d", got, DEFAULT_TIMEOUT_MS)
	}
	if got := options.Retries(); got != DEFAULT_RETRIES {
		t.Errorf("Retries() = %d, want %d", got, DEFAULT_RETRIES)
	}
	if got := options.PacketSize(); got != DEFAULT_PACKET_SIZE {
		t.Errorf("PacketSize() = %d, want %d", got, DEFAULT_PACKET_SIZE)
	}

	options.SetPort(12345)
	options.SetMaxHops(12)
	options.SetFirstHop(3)
	options.SetTimeoutMs(250)
	options.SetRetries(5)
	options.SetPacketSize(128)

	if options.Port() != 12345 || options.MaxHops() != 12 || options.FirstHop() != 3 ||
		options.TimeoutMs() != 250 || options.Retries() != 5 || options.PacketSize() != 128 {
		t.Fatal("option setters did not update their values")
	}
}

func TestDestAddrReturnsIPv4Address(t *testing.T) {
	got, err := destAddr("127.0.0.1")
	if err != nil {
		t.Fatalf("destAddr returned error: %v", err)
	}

	want := [4]byte{127, 0, 0, 1}
	if got != want {
		t.Errorf("destAddr = %v, want %v", got, want)
	}
}

func TestDestAddrRejectsUnknownHost(t *testing.T) {
	_, err := destAddr("invalid.invalid")
	if err == nil {
		t.Fatal("destAddr returned nil error for an unknown host")
	}
}

func TestTracerouteRejectsUnknownDestination(t *testing.T) {
	notifications := make(chan TracerouteHop)
	result, err := Traceroute("invalid.invalid", nil, notifications)
	if err == nil {
		t.Fatal("Traceroute returned nil error for an unknown destination")
	}
	if result.Hops == nil {
		t.Fatal("Traceroute returned nil hops for an invalid destination")
	}
	if len(result.Hops) != 0 {
		t.Fatalf("Traceroute returned %d hops, want 0", len(result.Hops))
	}
	if _, ok := <-notifications; ok {
		t.Fatal("Traceroute left the notification channel open")
	}
}

func TestTracerouteHopFormatting(t *testing.T) {
	hop := TracerouteHop{Address: [4]byte{192, 0, 2, 7}}
	if got := hop.AddressString(); got != "192.0.2.7" {
		t.Errorf("AddressString() = %q, want %q", got, "192.0.2.7")
	}
	if got := hop.HostOrAddressString(); got != "192.0.2.7" {
		t.Errorf("HostOrAddressString() = %q, want address", got)
	}

	hop.Host = "router.example"
	if got := hop.HostOrAddressString(); got != "router.example" {
		t.Errorf("HostOrAddressString() = %q, want hostname", got)
	}
}

func TestNotifySendsHopToAllChannels(t *testing.T) {
	first := make(chan TracerouteHop, 1)
	second := make(chan TracerouteHop, 1)
	want := TracerouteHop{Success: true, TTL: 2, ElapsedTime: time.Millisecond}

	notify(want, []chan TracerouteHop{first, second})

	if got := <-first; got != want {
		t.Errorf("first channel received %+v, want %+v", got, want)
	}
	if got := <-second; got != want {
		t.Errorf("second channel received %+v, want %+v", got, want)
	}
}

func TestCloseNotifyClosesAllChannels(t *testing.T) {
	first := make(chan TracerouteHop)
	second := make(chan TracerouteHop)

	closeNotify([]chan TracerouteHop{first, second})

	if _, ok := <-first; ok {
		t.Error("first channel is still open")
	}
	if _, ok := <-second; ok {
		t.Error("second channel is still open")
	}
}

func TestSocketAddr(t *testing.T) {
	result, err := socketAddr()
	if err != nil {
		t.Skipf("no non-loopback IPv4 interface available: %v", err)
	}
	if result == [4]byte{} {
		t.Error("socketAddr returned an empty IPv4 address")
	}
}

func TestTracerouteBaidu(t *testing.T) {
	options := &TracerouteOptions{}
	options.SetMaxHops(32)
	options.SetRetries(2)
	options.SetTimeoutMs(2000)
	notifications := make(chan TracerouteHop, 32)

	result, err := Traceroute("www.baidu.com", options, notifications)
	if err != nil {
		t.Skipf("traceroute unavailable in this environment: %v", err)
	}
	if len(result.Hops) == 0 {
		t.Fatal("Traceroute returned no hops for www.baidu.com")
	}

	wantDestination, err := destAddr("www.baidu.com")
	if err != nil {
		t.Fatalf("destAddr(www.baidu.com) returned error: %v", err)
	}
	if result.DestinationAddress != wantDestination {
		t.Errorf("destination = %v, want %v", result.DestinationAddress, wantDestination)
	}
}

func TestMain(t *testing.T) {
	options := &TracerouteOptions{}
	options.SetMaxHops(16)
	options.SetRetries(2)
	options.SetTimeoutMs(500)

	hops := make(chan TracerouteHop, options.MaxHops())
	result, err := Traceroute("www.baidu.com", options, hops)
	if err != nil {
		t.Skipf("traceroute unavailable in this environment: %v", err)
		return
	}

	for hop := range hops {
		if hop.Success {
			fmt.Printf("%d  %-39s  %s\n", hop.TTL, hop.HostOrAddressString(), hop.ElapsedTime)
			continue
		}
		fmt.Printf("%d  *\n", hop.TTL)
	}

	fmt.Printf("destination: %s (%d hops)\n", net.IP(result.DestinationAddress[:]), len(result.Hops))

}
