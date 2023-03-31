package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"
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
	return string(bytePassword)
}

const domain = "b.com"
const configDir = "$HOME/.buoy"
const resolverPort = 8053

func main() {
	setup := flag.Bool("setup", false, "setup the buoy config directory and certs")
	flag.Parse()

	if *setup {
		sudo := os.Getenv("SUDO_USER")
		if sudo == "" {
			fmt.Println("\nyou must run this command with sudo")
			os.Exit(1)
		}

		uid, uidError := strconv.Atoi(os.Getenv("SUDO_UID"))
		gid, gidError := strconv.Atoi(os.Getenv("SUDO_GID"))
		if uidError != nil || gidError != nil {
			fmt.Println("\nerror getting sudo user id")
			os.Exit(1)
		}

		dnsError := buoy.SetupDNSResolver(uid, gid, domain, resolverPort)
		if dnsError != nil {
			fmt.Printf("\nerror setting up dns resolver: %s\n", dnsError.Error())
			os.Exit(1)
		}

		configDir, configDirError := buoy.SetupConfigDir(uid, gid, configDir)
		if configDirError != nil {
			fmt.Printf("\nerror setting up config dir: %s\n", configDirError.Error())
			os.Exit(1)
		}

		certError := buoy.SetupCert(uid, gid, configDir, domain)
		if certError != nil {
			fmt.Printf("\nerror setting up ssl cert: %s\n", certError.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	serviceArgs := flag.Args()
	services, parseServicesError := buoy.ParseServices(domain, serviceArgs)
	if parseServicesError != nil {
		fmt.Printf(
			"\nerror parsing services: %s\n  a service should be in the form of subdomain/path:port\n",
			parseServicesError.Error(),
		)
		os.Exit(1)
	}

	configDir, configDirError := buoy.GetConfigDir(configDir)
	if configDirError != nil {
		fmt.Printf("\nerror getting config dir: %s\n", configDirError.Error())
		os.Exit(1)
	}

	certFile, keyFile, certError := buoy.GetCert(configDir, domain)
	if certError != nil {
		fmt.Printf("\nerror getting ssl cert: %s\n", certError.Error())
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		dnsError := buoy.StartDNSResolver(resolverPort)
		if dnsError != nil {
			fmt.Printf("\nerror starting dns resolver: %s\n", dnsError.Error())
			os.Exit(1)
		}
	}()
	go func() {
		defer wg.Done()
		rproxyError := buoy.StartReverseProxy(services, certFile, keyFile)
		if rproxyError != nil {
			fmt.Printf("\nerror starting reverse proxy: %s\n", rproxyError.Error())
			os.Exit(1)
		}
	}()

	wg.Wait()
}
