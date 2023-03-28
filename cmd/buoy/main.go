package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tinyprint/buoy/internal/buoy"
)

func main() {
	domain := flag.String("domain", "b.com", "domain to use for buoy")
	flag.Parse()

	serviceArgs := flag.Args()
	services, parseServicesError := buoy.ParseServices(serviceArgs)
	if parseServicesError != nil {
		fmt.Printf(
			"error parsing services: %s\n  a service should be in the form of subdomain/path:port\n",
			parseServicesError.Error(),
		)
		os.Exit(1)
	}

	fmt.Printf("buoy domain: %v\n", *domain)
	fmt.Printf("args: %v\n", serviceArgs)
	fmt.Printf("services: %v\n", services)

	go buoy.StartDNSResolver()
	buoy.StartReverseProxy(*domain, services, "cert.pem", "key.pem")
}
