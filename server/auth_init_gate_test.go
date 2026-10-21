package server

import (
	"net/http/httptest"
	"testing"
)

// A fresh install exposed on a network must not be claimable by whoever reaches
// POST /api/auth/init first; only the operator (loopback, or holder of the
// bootstrap token) may set the admin password.
func TestInitAllowedRejectsRemoteFirstVisitor(t *testing.T) {
	for _, tc := range []struct {
		name, envToken, remote, header string
		want                           bool
	}{
		{"remote without token config", "", "203.0.113.7:4000", "", false},
		{"docker bridge without token config", "", "172.17.0.1:4000", "", false},
		{"ipv4 loopback without token config", "", "127.0.0.1:4000", "", true},
		{"ipv6 loopback without token config", "", "[::1]:4000", "", true},
		{"unparseable remote address", "", "garbage", "", false},
		{"remote with matching token", "s3cret-bootstrap", "203.0.113.7:4000", "s3cret-bootstrap", true},
		{"remote with wrong token", "s3cret-bootstrap", "203.0.113.7:4000", "guess", false},
		{"remote with missing token", "s3cret-bootstrap", "203.0.113.7:4000", "", false},
		// Configured token is authoritative: loopback alone is not enough once set,
		// because a reverse proxy on the same host would make every visitor loopback.
		{"loopback with token configured but not sent", "s3cret-bootstrap", "127.0.0.1:4000", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(initTokenEnv, tc.envToken)
			r := httptest.NewRequest("POST", "/api/auth/init", nil)
			r.RemoteAddr = tc.remote
			if tc.header != "" {
				r.Header.Set("X-Init-Token", tc.header)
			}
			if got := initAllowed(r); got != tc.want {
				t.Fatalf("initAllowed=%v want %v", got, tc.want)
			}
		})
	}
}
