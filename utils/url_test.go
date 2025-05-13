package utils

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"
)

func TestFullRequestURL(t *testing.T) {
	tests := []struct {
		name     string
		request  *http.Request
		expected string
	}{
		{
			name: "HTTP request without TLS",
			request: &http.Request{
				Host: "example.com",
				URL: &url.URL{
					Path:     "/test",
					RawQuery: "param=value",
					Fragment: "section",
				},
			},
			expected: "http://example.com/test?param=value#section",
		},
		{
			name: "HTTPS request with TLS",
			request: &http.Request{
				Host: "example.com",
				URL: &url.URL{
					Path: "/secure",
				},
				TLS: &tls.ConnectionState{},
			},
			expected: "https://example.com/secure",
		},
		{
			name: "Request with X-Forwarded-Proto header",
			request: &http.Request{
				Host: "example.com",
				URL: &url.URL{
					Path: "/forwarded",
				},
				Header: http.Header{
					"X-Forwarded-Proto": []string{"https"},
				},
			},
			expected: "https://example.com/forwarded",
		},
		{
			name: "Request with X-Forwarded-Proto header and TLS",
			request: &http.Request{
				Host: "example.com",
				URL: &url.URL{
					Path: "/conflict",
				},
				Header: http.Header{
					"X-Forwarded-Proto": []string{"http"},
				},
				TLS: &tls.ConnectionState{},
			},
			expected: "http://example.com/conflict", // X-Forwarded-Proto should take precedence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FullRequestURL(tt.request)
			if got.String() != tt.expected {
				t.Errorf("FullRequestURL() = %v, want %v", got.String(), tt.expected)
			}
		})
	}
}
