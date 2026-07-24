package ratelimit

import (
	"net/http/httptest"
	"testing"
)

// ensures that ClientIP correctly extracts IPv4 and IPv6 addresses from the RemoteAddr header
func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       string
	}{
		{
			name:       "IPv4 with port",
			remoteAddr: "192.0.2.10:54321",
			want:       "192.0.2.10",
		},
		{
			name:       "IPv6 with port",
			remoteAddr: "[2001:db8::1]:54321",
			want:       "2001:db8::1",
		},
		{
			name:       "IPv4 without port",
			remoteAddr: "192.0.2.10",
			want:       "192.0.2.10",
		},
		{
			name:       "empty remote address",
			remoteAddr: "",
			want:       "",
		},
		{
			name:       "trims whitespace",
			remoteAddr: " 192.0.2.10:54321 ",
			want:       "192.0.2.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			got := ClientIP(req)

			if got != tt.want {
				t.Fatalf(
					"ClientIP() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

// ensures that ClientIP handles nil requests safely
func TestClientIPWithNilRequest(t *testing.T) {
	if got := ClientIP(nil); got != "" {
		t.Fatalf("ClientIP(nil) = %q, want empty string", got)
	}
}

// ensures that ClientIP ignores forwarded headers and always uses RemoteAddr
func TestClientIPIgnoresForwardedHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.0.67:12345"

	req.Header.Set("X-Forwarded-For", "192.0.0.68")
	req.Header.Set("X-Real-IP", "192.0.0.68")

	got := ClientIP(req)

	if got != "192.0.0.67" {
		t.Fatalf("ClientIP() = %q, want RemoteAddr IP", got)
	}
}