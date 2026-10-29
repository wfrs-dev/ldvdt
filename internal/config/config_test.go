package config

import "testing"

// The trainer serves an unauthenticated SPA, so it must never bind a public
// interface: an empty listen host makes Echo listen on every address.
func TestDefaultListenAddrIsLoopback(t *testing.T) {
	c := Default()

	if got, want := c.Host, "127.0.0.1"; got != want {
		t.Fatalf("Host = %q, want %q", got, want)
	}
	if got, want := c.ListenAddr(), "127.0.0.1:10987"; got != want {
		t.Fatalf("ListenAddr() = %q, want %q", got, want)
	}
}

func TestListenAddrFollowsConfiguredPort(t *testing.T) {
	c := Default()
	c.Port = 12345

	if got, want := c.ListenAddr(), "127.0.0.1:12345"; got != want {
		t.Fatalf("ListenAddr() = %q, want %q", got, want)
	}
}
