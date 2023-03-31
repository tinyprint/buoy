package buoy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	proxies = map[string]*httputil.ReverseProxy{}
)

type reverseProxyHandler struct {
	serviceMap map[string]int
}

func buildServiceMap(services []Service) map[string]int {
	serviceMap := map[string]int{}
	for _, service := range services {
		serviceID := fmt.Sprintf("%s.%s/%s", service.Subdomain, service.Domain, service.Path)
		serviceMap[serviceID] = service.Port
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
	fmt.Printf("# reverse-proxy - received request host=%s path=%s\n", r.Host, r.URL.Path)

	serviceID := getServiceID(r.Host, r.URL.Path)

	if _, ok := h.serviceMap[serviceID]; !ok {
		serviceID = getServiceID(r.Host, "/")
	}

	if _, ok := h.serviceMap[serviceID]; !ok {
		fmt.Fprintf(w, "buoy - nothing to see here")
		return
	}

	if proxy, ok := proxies[serviceID]; ok {
		proxy.ServeHTTP(w, r)
		return
	}

	servicePort := h.serviceMap[serviceID]
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%d", servicePort),
	})
	proxies[serviceID] = proxy
	proxy.ServeHTTP(w, r)
}

// StartReverseProxy starts the reverse proxy
func StartReverseProxy(services []Service, certFile string, keyFile string) error {
	fmt.Println("# reverse-proxy - starting...")

	serviceMap := buildServiceMap(services)

	h := &reverseProxyHandler{
		serviceMap: serviceMap,
	}
	reverseProxyServer := &http.Server{
		Addr:    ":443",
		Handler: h,
	}
	err := reverseProxyServer.ListenAndServeTLS(certFile, keyFile)
	return fmt.Errorf("error listening and serving tls: %s", err)
}
