package rproxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/tinyprint/buoy/internal/pkg/config"
)

var (
	proxies = map[string]*httputil.ReverseProxy{}
)

type reverseProxyHandler struct {
	config     config.Config
	serviceMap map[string]int
}

func buildServiceMap(config config.Config) map[string]int {
	serviceMap := map[string]int{}
	for serviceGroupID, serviceGroup := range config.Services {
		for servicePartialID, service := range serviceGroup {
			serviceID := fmt.Sprintf("%s.%s%s", serviceGroupID, config.RootDomain, servicePartialID)
			serviceMap[serviceID] = service.Port
		}
	}
	return serviceMap
}

func getServiceID(host string, path string) string {
	if path == "/" {
		return host + "/"
	}

	if path[0] == '/' {
		path = path[1:]
	}
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			return host + "/" + path[:i]
		}
	}
	return host + "/" + path
}

func (h *reverseProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("received request host=%s path=%s\n", r.Host, r.URL.Path)

	serviceID := getServiceID(r.Host, r.URL.Path)

	if proxy, ok := proxies[serviceID]; ok {
		proxy.ServeHTTP(w, r)
		return
	}

	if servicePort, ok := h.serviceMap[serviceID]; ok {
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%d", servicePort),
		})
		proxies[serviceID] = proxy
		proxy.ServeHTTP(w, r)
		return
	}

	fmt.Fprintf(w, "nothing to see here")
}

// StartReverseProxy starts the reverse proxy
func StartReverseProxy(config config.Config) error {
	fmt.Println("starting Buoy reverse proxy")

	serviceMap := buildServiceMap(config)

	h := &reverseProxyHandler{
		config:     config,
		serviceMap: serviceMap,
	}
	reverseProxyServer := &http.Server{
		Addr:    ":443",
		Handler: h,
	}
	err := reverseProxyServer.ListenAndServeTLS(config.GetSSLCertPath(), config.GetSSLKeyPath())
	return fmt.Errorf("error listening and serving TLS: %s", err)
}
