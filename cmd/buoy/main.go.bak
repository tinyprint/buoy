package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/tinyprint/buoy/internal/pkg/config"
	"github.com/tinyprint/buoy/internal/pkg/dns"
	"github.com/tinyprint/buoy/internal/pkg/rproxy"
)

func startDNSResolver() {
	fmt.Println("starting Buoy DNS resolver")
	p, err := net.ListenPacket("udp", ":8053")
	if err != nil {
		fmt.Printf("error starting DNS resolver: %s", err)
		os.Exit(1)
	}
	defer p.Close()

	for {
		buf := make([]byte, 512)
		n, addr, err := p.ReadFrom(buf)
		fmt.Printf("connection [%s]...\n", addr.String())
		if err != nil {
			fmt.Printf("connection error [%s]: %s\n", addr.String(), err)
			continue
		}
		go dns.HandlePacket(p, addr, buf[:n])
	}
}

var (
	configDir = flag.String(
		"config-dir",
		"$HOME/.buoy",
		"Directory to store Buoy configuration",
	)
	initialize = flag.Bool(
		"init",
		false,
		"Initialize Buoy configuration and SSL Certs",
	)
)

func main() {
	flag.Parse()

	if *initialize == true {
		fmt.Println("initializing Buoy configuration and SSL Certs")
		config.CreateConfig(*configDir)

		// TODO: Create SSL Certs and add them to keychain
	}

	config, err := config.LoadConfig(*configDir)
	if err != nil {
		fmt.Printf("error loading Buoy configuration: %s\n", err)
		os.Exit(1)
	}

	// Lightster says to use a channel
	go startDNSResolver()
	err = rproxy.StartReverseProxy(config)
	if err != nil {
		fmt.Printf("error starting Buoy reverse proxy: %s\n", err)
		os.Exit(1)
	}
}
