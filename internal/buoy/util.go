package buoy

import "strings"

// DomainToFilename gets a nice filename for the given domain
func DomainToFilename(domain string) string {
	return strings.ReplaceAll(domain, ".", "-")
}
