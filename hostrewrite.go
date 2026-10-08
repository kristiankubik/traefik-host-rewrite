package traefik_host_rewrite

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type Config struct {
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`
	Mode   string `json:"mode,omitempty"`  // suffix|replace
	Debug  bool   `json:"debug,omitempty"` // true|false
}

func CreateConfig() *Config {
	return &Config{}
}

// HostRewrite rewrites request Host values using exact or DNS suffix replacement.
type HostRewrite struct {
	next   http.Handler
	source string
	target string
	mode   string
	debug  bool
}

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

	mode := strings.ToLower(strings.TrimSpace(config.Mode))
	if mode == "" {
		mode = "suffix"
	}
	switch mode {
	case "suffix", "replace":
	default:
		return nil, fmt.Errorf(
			"invalid mode %q: expected suffix or replace",
			mode,
		)
	}

	return &HostRewrite{
		next:   next,
		source: source,
		target: target,
		mode:   mode,
		debug:  config.Debug,
	}, nil
}

func (m *HostRewrite) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	original := req.Host
	host := stripPort(req.Host)
	normalizedHost := strings.ToLower(strings.TrimSuffix(host, "."))

	switch m.mode {
	case "replace":
		if normalizedHost == m.source {
			req.Host = m.target
		}
	case "suffix":
		switch {
		case normalizedHost == m.source:
			req.Host = m.target
		case strings.HasSuffix(normalizedHost, "."+m.source):
			prefix := strings.TrimSuffix(normalizedHost, "."+m.source)
			req.Host = prefix + "." + m.target
		}
	}

	if m.debug && original != req.Host {
		fmt.Printf(
			"hostrewrite: mode=%q source=%q target=%q host=%q -> %q\n",
			m.mode,
			m.source,
			m.target,
			original,
			req.Host,
		)
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
	return hostport
}
