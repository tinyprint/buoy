package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/tinyprint/buoy/internal/buoy"
	"golang.org/x/term"
)

func getPassword() string {
	fmt.Print("Enter password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return ""
	}
	fmt.Println("")
	return string(bytePassword)
}

func main() {
	domain := flag.String("domain", "b.com", "domain to use for buoy")
	configDirArg := flag.String(
		"config-dir",
		"$HOME/.buoy",
		"directory to store config files and certs",
	)
	flag.Parse()

	serviceArgs := flag.Args()
	services, parseServicesError := buoy.ParseServices(*domain, serviceArgs)
	if parseServicesError != nil {
		fmt.Printf(
			"error parsing services: %s\n  a service should be in the form of subdomain/path:port\n",
			parseServicesError.Error(),
		)
		os.Exit(1)
	}

	configDir, configDirError := buoy.GetOrCreateConfigDir(*configDirArg)
	if configDirError != nil {
		fmt.Printf("error getting config dir: %s\n", configDirError.Error())
		os.Exit(1)
	}

	certFile, keyFile, certError := buoy.GetOrCreateCert(configDir, *domain, getPassword)
	if certError != nil {
		fmt.Printf("error getting ssl cert: %s\n", certError.Error())
		os.Exit(1)
	}

	go buoy.StartDNSResolver()
	buoy.StartReverseProxy(services, certFile, keyFile)
}
