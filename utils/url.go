package utils

import (
	"net/http"
	"net/url"
)

func FullRequestURL(r *http.Request) *url.URL {
	scheme := "http"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if r.TLS != nil {
		scheme = "https"
	}

	return &url.URL{
		Scheme:   scheme,
		Host:     r.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
		Fragment: r.URL.Fragment,
	}
}
