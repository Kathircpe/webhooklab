package service

import (
	"context"
	"net"
	"testing"
)

func TestIPBlocked(t *testing.T) {
	cases := []struct {
		ip           string
		allowPrivate bool
		want         bool
	}{
		{"127.0.0.1", false, true},
		{"127.0.0.1", true, false},
		{"10.0.0.5", false, true},
		{"10.0.0.5", true, false},
		{"169.254.169.254", false, true},
		{"169.254.169.254", true, true},
		{"8.8.8.8", false, false},
		{"::1", false, true},
		{"fd00::1", false, true},
	}
	for _, c := range cases {
		if got := ipBlocked(net.ParseIP(c.ip), c.allowPrivate); got != c.want {
			t.Errorf("ipBlocked(%s, %v) = %v, want %v", c.ip, c.allowPrivate, got, c.want)
		}
	}
}

func TestValidateDestinationBlocksSSRF(t *testing.T) {
	ctx := context.Background()
	for _, u := range []string{
		"http://127.0.0.1:8080/x",
		"http://169.254.169.254/latest/meta-data",
		"http://localhost/x",
		"http://[::1]/x",
		"ftp://example.com/x",
	} {
		if _, err := validateDestination(ctx, u, false); err == nil {
			t.Errorf("expected %s to be blocked", u)
		}
	}
	// Metadata endpoints stay blocked even when private destinations are enabled.
	for _, u := range []string{
		"http://169.254.169.254/latest/meta-data",
		"http://metadata.google.internal/x",
	} {
		if _, err := validateDestination(ctx, u, true); err == nil {
			t.Errorf("expected %s to be blocked with allowPrivate", u)
		}
	}
}
