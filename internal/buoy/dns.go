package buoy

import (
	"fmt"
	"net"
	"os"

	"golang.org/x/net/dns/dnsmessage"
)

const resolverDir = "/etc/resolver"

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
func StartDNSResolver(resolverPort int) error {
	fmt.Println("# dns-resolver - starting...")
	p, listenError := net.ListenPacket("udp", fmt.Sprintf(":%d", resolverPort))
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

// SetupDNSResolver sets up the DNS resolver
func SetupDNSResolver(uid int, gid int, domain string, resolverPort int) error {
	fmt.Printf("# creating resolver directory %s...", resolverDir)

	mkdirError := os.MkdirAll(resolverDir, 0755)
	if mkdirError != nil {
		return fmt.Errorf("error creating resolver directory: %s", mkdirError)
	}

	chownError := os.Chown(resolverDir, uid, gid)
	if chownError != nil {
		return fmt.Errorf("error changing resolver directory owner: %s", chownError)
	}

	fmt.Println("done")

	resolverFilePath := resolverDir + "/" + domain
	resolverTmpl := fmt.Sprintf(`nameserver 127.0.0.1
port %d
`, resolverPort)

	fmt.Printf("# creating resolver file %s...", resolverFilePath)

	// create resolver file if it doesn't exist
	resolverFile, openError := os.OpenFile(resolverFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if openError != nil {
		return fmt.Errorf("error opening/creating resolver file %s: %s", resolverFilePath, openError)
	}

	_, writeError := resolverFile.WriteString(resolverTmpl)
	if writeError != nil {
		return fmt.Errorf("error writing resolver file %s: %s", resolverFile.Name(), writeError)
	}

	if err := resolverFile.Close(); err != nil {
		return fmt.Errorf("error closing %s: %s", resolverFile.Name(), err)
	}

	chownFileError := os.Chown(resolverFilePath, uid, gid)
	if chownFileError != nil {
		return fmt.Errorf("error changing resolver file owner %s: %s", resolverFilePath, chownFileError)
	}

	fmt.Println("done")

	return nil
}
