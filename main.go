/*IPv4 traceroute (UDP probe only) in Go*/
package main

import (
	geoip "./telize"
	ipv4 "code.google.com/p/go.net/ipv4"
	"flag"
	"fmt"
	log "github.com/dmuth/google-go-log4go"
	"net"
	"syscall"
)

const (
	MAX_HOP     = 32
	TIME_OUT_MS = 3000
)

type Hop_ret struct {
	Addr    net.IP
	TTL     int
	success bool
}

func (h *Hop_ret) String() string {
	return "fuckit"
}

func (h *Hop_ret) Success() bool {
	if h.success {
		return true
	} else {
		return false
	}
}

func main() {
	flag.Parse()
	ip_addr := net.ParseIP(flag.Arg(0))

	fmt.Printf("Starting traceroute for ip = %s\n", flag.Arg(0))
	for i := 0; i < MAX_HOP; i++ {
		resHop, err := Hop(1337+i, i+1, ip_addr)
		if err != nil {
			log.Errorf("[TTL: %d] %q", i+1, err)
		}
		if resHop.Success() {
			GeoIp := geoip.TelizeRequest{IP: resHop.Addr}
			result, err := GeoIp.GetGeo()
			if err != nil {
				log.Errorf("[TTL: %d] %q", i+1, err)
			}
			if result.Type == geoip.RES_SUCCESS {
				fmt.Print(resHop.TTL, " \t ", resHop.Addr.String(), " \t [", result.GeoInfo.Country, " ", result.GeoInfo.City, " ", result.GeoInfo.Asn, "]\n")
			} else {
				fmt.Print(resHop.TTL, " \t ", resHop.Addr.String(), "  \t [",  result.Error.Message, "]\n")
			}
		} else {
			fmt.Print(resHop.TTL, " \t * * *\n")
		}
	}

}

func Hop(port, ttl int, IP_addr net.IP) (*Hop_ret, error) {

	ret_addr := net.IPv4(0, 0, 0, 0)
	success := false
	// make sockets
	send_udp_s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		return nil, err
	}
	recv_icmp_s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		return nil, err
	}

	//editing TTL value for outgoing IPv4 packets
	if err := syscall.SetsockoptInt(send_udp_s, syscall.SOL_IP, syscall.IP_TTL, ttl); err != nil {
		return nil, err
	}
	tv := syscall.NsecToTimeval(1000 * 1000 * TIME_OUT_MS)
	syscall.SetsockoptTimeval(recv_icmp_s, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)

	defer syscall.Close(send_udp_s)
	defer syscall.Close(recv_icmp_s)

	//connect sockets
	if err := syscall.Bind(recv_icmp_s, &syscall.SockaddrInet4{Port: port, Addr: [4]byte{137, 224, 226, 47}}); err != nil {
		return nil, err
	}

	//send udp-packet
	var IP [4]byte
	copy(IP[:], IP_addr.To4())
	if err := syscall.Sendto(send_udp_s, []byte{0x42, 0x42}, 0, &syscall.SockaddrInet4{Port: 1337, Addr: IP}); err != nil {
		return nil, err
	}

	//receive ICMP
	recv_buffer := make([]byte, 4096)
	_, _, err = syscall.Recvfrom(recv_icmp_s, recv_buffer, 0)
	if err == nil {
		header, err := ipv4.ParseHeader(recv_buffer)
		if err != nil {
			log.Errorf("%q", err)
		}
		success = true
		ret_addr = header.Src
	} else {
		//time out
		success = false
		ret_addr = net.IPv4(0, 0, 0, 0)
		//log.Errorf("%q", err)
	}
	//resolve (timeout) errors, retry or return false...
	return &Hop_ret{Addr: ret_addr, TTL: ttl, success: success}, nil
}
