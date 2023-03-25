package main

import (
	"flag"
	"fmt"

	"github.com/tinyprint/buoy/internal/pkg/config"
)

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
	}
}
