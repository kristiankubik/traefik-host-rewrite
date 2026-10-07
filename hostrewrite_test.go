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
		host     string
		expected string
	}{
		{
			name:     "root domain",
			host:     "oni.tld.example.com",
			expected: "oni.internal",
		},
		{
			name:     "subdomain",
			host:     "conntest.oni.tld.example.com",
			expected: "conntest.oni.internal",
		},
		{
			name:     "nested subdomain",
			host:     "foo.bar.oni.tld.example.com",
			expected: "foo.bar.oni.internal",
		},
		{
			name:     "host with port",
			host:     "conntest.oni.tld.example.com:443",
			expected: "conntest.oni.internal",
		},
		{
			name:     "case insensitive",
			host:     "CONNTEST.ONI.tld.example.com",
			expected: "conntest.oni.internal",
		},
		{
			name:     "unrelated host",
			host:     "example.com",
			expected: "example.com",
		},
		{
			name:     "suffix lookalike",
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