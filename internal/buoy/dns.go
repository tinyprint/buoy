package buoy

import (
	"fmt"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

func handlePacket(pc net.PacketConn, addr net.Addr, buf []byte) error {
	p := dnsmessage.Parser{}
	header, err := p.Start(buf)
	if err != nil {
		return err
	}
	question, err := p.Question()
	if err != nil {
		return err
	}

	response := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:            header.ID,
			Response:      true,
			Authoritative: true,
		},
		Questions: []dnsmessage.Question{
			question,
		},
		Answers: []dnsmessage.Resource{
			{
				Header: dnsmessage.ResourceHeader{
					Name:  question.Name,
					Type:  dnsmessage.TypeA,
					Class: dnsmessage.ClassINET,
				},
				Body: &dnsmessage.AResource{
					A: [4]byte{127, 0, 0, 1},
				},
			},
		},
	}
	responseBuffer, err := response.Pack()
	if err != nil {
		return err
	}
	_, err = pc.WriteTo(responseBuffer, addr)
	if err != nil {
		return err
	}

	return nil
}

func handle(pc net.PacketConn, addr net.Addr, buf []byte) {
	if err := handlePacket(pc, addr, buf); err != nil {
		fmt.Printf("# dns-resolver - could not handle packet [%s]: %s\n", addr.String(), err)
	}
}

// StartDNSResolver starts the DNS resolver
func StartDNSResolver() error {
	fmt.Println("# dns-resolver - starting...")
	p, listenError := net.ListenPacket("udp", ":8053")
	if listenError != nil {
		return fmt.Errorf("error starting dns-resolver: %s", listenError)
	}
	defer p.Close()
	fmt.Println("# dns-resolver - started")

	for {
		buf := make([]byte, 512)
		n, addr, readError := p.ReadFrom(buf)
		fmt.Printf("# dns-resolver - connection [%s]...\n", addr.String())
		if readError != nil {
			fmt.Printf("# dns-resolver - connection error [%s]: %s\n", addr.String(), readError)
			continue
		}
		go handle(p, addr, buf[:n])
	}
}
