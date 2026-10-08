package traefik_host_rewrite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHostRewrite(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		host     string
		expected string
	}{
		// Default mode should remain suffix for backwards compatibility.
		{
			name:     "default mode root domain",
			host:     "oni.tld.example.com",
			expected: "oni.internal",
		},
		{
			name:     "default mode subdomain",
			host:     "conntest.oni.tld.example.com",
			expected: "conntest.oni.internal",
		},

		// Suffix mode.
		{
			name:     "suffix root domain",
			mode:     "suffix",
			host:     "oni.tld.example.com",
			expected: "oni.internal",
		},
		{
			name:     "suffix subdomain",
			mode:     "suffix",
			host:     "conntest.oni.tld.example.com",
			expected: "conntest.oni.internal",
		},
		{
			name:     "suffix nested subdomain",
			mode:     "suffix",
			host:     "foo.bar.oni.tld.example.com",
			expected: "foo.bar.oni.internal",
		},
		{
			name:     "suffix host with port",
			mode:     "suffix",
			host:     "conntest.oni.tld.example.com:443",
			expected: "conntest.oni.internal",
		},
		{
			name:     "suffix case insensitive",
			mode:     "suffix",
			host:     "CONNTEST.ONI.tld.example.com",
			expected: "conntest.oni.internal",
		},
		{
			name:     "suffix unrelated host",
			mode:     "suffix",
			host:     "example.com",
			expected: "example.com",
		},
		{
			name:     "suffix lookalike",
			mode:     "suffix",
			host:     "eviloni.tld.example.com",
			expected: "eviloni.tld.example.com",
		},

		// Replace mode.
		{
			name:     "replace exact host",
			mode:     "replace",
			host:     "oni.tld.example.com",
			expected: "oni.internal",
		},
		{
			name:     "replace exact host with port",
			mode:     "replace",
			host:     "oni.tld.example.com:443",
			expected: "oni.internal",
		},
		{
			name:     "replace case insensitive",
			mode:     "replace",
			host:     "ONI.TLD.EXAMPLE.COM",
			expected: "oni.internal",
		},
		{
			name:     "replace does not match subdomain",
			mode:     "replace",
			host:     "conntest.oni.tld.example.com",
			expected: "conntest.oni.tld.example.com",
		},
		{
			name:     "replace does not match nested subdomain",
			mode:     "replace",
			host:     "foo.bar.oni.tld.example.com",
			expected: "foo.bar.oni.tld.example.com",
		},
		{
			name:     "replace unrelated host",
			mode:     "replace",
			host:     "example.com",
			expected: "example.com",
		},
		{
			name:     "replace suffix lookalike",
			mode:     "replace",
			host:     "eviloni.tld.example.com",
			expected: "eviloni.tld.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedHost string

			next := http.HandlerFunc(func(
				rw http.ResponseWriter,
				req *http.Request,
			) {
				receivedHost = req.Host
				rw.WriteHeader(http.StatusOK)
			})

			handler, err := New(
				context.Background(),
				next,
				&Config{
					Source: "oni.tld.example.com",
					Target: "oni.internal",
					Mode:   tt.mode,
				},
				"host-rewrite",
			)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(
				http.MethodGet,
				"http://example/",
				nil,
			)
			req.Host = tt.host

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if receivedHost != tt.expected {
				t.Errorf(
					"expected host %q, got %q",
					tt.expected,
					receivedHost,
				)
			}
		})
	}
}

func TestHostRewriteInvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "empty source",
			config: Config{
				Target: "oni.internal",
			},
		},
		{
			name: "empty target",
			config: Config{
				Source: "oni.tld.example.com",
			},
		},
		{
			name: "same source and target",
			config: Config{
				Source: "oni.internal",
				Target: "oni.internal",
			},
		},
		{
			name: "invalid mode",
			config: Config{
				Source: "oni.tld.example.com",
				Target: "oni.internal",
				Mode:   "regex",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(
				http.ResponseWriter,
				*http.Request,
			) {
			})

			_, err := New(
				context.Background(),
				next,
				&tt.config,
				"host-rewrite",
			)

			if err == nil {
				t.Fatal("expected configuration error, got nil")
			}
		})
	}
}
