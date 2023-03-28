package buoy

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	// Service is a service to be proxied
	Service struct {
		Subdomain string
		Path      string
		Port      int
	}
)

// ParseServices parses a list of services from a list of strings
func ParseServices(serviceArgs []string) ([]Service, error) {
	services := make([]Service, len(serviceArgs))
	for i, serviceArg := range serviceArgs {
		servicePieces := strings.Split(serviceArg, ":")
		if len(servicePieces) != 2 {
			return nil, fmt.Errorf("invalid service '%s'", serviceArg)
		}

		portStr := servicePieces[1]
		port, portConvError := strconv.Atoi(portStr)
		if portConvError != nil {
			return nil, fmt.Errorf("invalid port '%v' for service '%v'", portStr, serviceArg)
		}

		var subdomain, path string
		subdomainAndPath := strings.Split(servicePieces[0], "/")
		if len(subdomainAndPath) == 1 {
			subdomain = subdomainAndPath[0]
		} else if len(subdomainAndPath) == 2 {
			subdomain = subdomainAndPath[0]
			path = subdomainAndPath[1]
		} else {
			return nil, fmt.Errorf("invalid path (more than 1) for service '%s'", serviceArg)
		}

		services[i] = Service{
			Subdomain: subdomain,
			Path:      path,
			Port:      port,
		}
	}

	return services, nil
}
