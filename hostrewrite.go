package traefik_host_rewrite

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// Config defines the middleware configuration.
type Config struct {
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// HostRewrite rewrites request Host values from one DNS suffix to another.
type HostRewrite struct {
	next   http.Handler
	source string
	target string
}

// New creates a new HostRewrite middleware.
func New(
	_ context.Context,
	next http.Handler,
	config *Config,
	_ string,
) (http.Handler, error) {
	source := normalizeDomain(config.Source)
	target := normalizeDomain(config.Target)

	if source == "" {
		return nil, fmt.Errorf("source must not be empty")
	}

	if target == "" {
		return nil, fmt.Errorf("target must not be empty")
	}

	if source == target {
		return nil, fmt.Errorf("source and target must be different")
	}

	return &HostRewrite{
		next:   next,
		source: source,
		target: target,
	}, nil
}

func (m *HostRewrite) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	host := stripPort(req.Host)
	normalizedHost := strings.ToLower(strings.TrimSuffix(host, "."))

	switch {
	case normalizedHost == m.source:
		req.Host = m.target

	case strings.HasSuffix(normalizedHost, "."+m.source):
		prefix := strings.TrimSuffix(normalizedHost, "."+m.source)
		req.Host = prefix + "." + m.target
	}

	m.next.ServeHTTP(rw, req)
}

func normalizeDomain(domain string) string {
	return strings.ToLower(
		strings.TrimSuffix(strings.TrimSpace(domain), "."),
	)
}

func stripPort(hostport string) string {
	host, _, err := net.SplitHostPort(hostport)
	if err == nil {
		return host
	}
	// Normal hostname without a port.
	return hostport
}