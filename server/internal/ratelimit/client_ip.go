package ratelimit

import (
	"net"
	"net/http"
	"strings"
)

func ClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	remoteAddr := strings.TrimSpace(r.RemoteAddr)

	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}

	return remoteAddr
}