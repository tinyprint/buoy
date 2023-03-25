package dns

import (
	"fmt"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

// HandlePacket is the entry point for the resolver
func HandlePacket(pc net.PacketConn, addr net.Addr, buf []byte) {
	if err := handlePacket(pc, addr, buf); err != nil {
		fmt.Printf("dns resolver could not handle packet [%s]: %s\n", addr.String(), err)
	}
}

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
