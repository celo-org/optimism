package client

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// listening starts a TCP listener on loopback and returns its address.
func listening(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })
	return l.Addr().String()
}

// closedPort returns a loopback address that nothing is listening on.
func closedPort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())
	return addr
}

// withProxy swaps the environment proxy lookup for the duration of a test.
func withProxy(t *testing.T, proxy string) {
	t.Helper()
	prev := proxyForRequest
	t.Cleanup(func() { proxyForRequest = prev })
	if proxy == "" {
		proxyForRequest = func(*http.Request) (*url.URL, error) { return nil, nil }
		return
	}
	u, err := url.Parse(proxy)
	require.NoError(t, err)
	proxyForRequest = func(*http.Request) (*url.URL, error) { return u, nil }
}

func TestIsURLAvailable_NoProxy(t *testing.T) {
	withProxy(t, "")
	ctx := context.Background()

	require.True(t, IsURLAvailable(ctx, "http://"+listening(t), time.Second),
		"a reachable target should be available")
	require.False(t, IsURLAvailable(ctx, "http://"+closedPort(t), time.Second),
		"an unreachable target should not be available")
}

// The enclave case: the target is not directly reachable, but the proxy the RPC
// client will actually dial is. Before this was proxy-aware the check probed the
// target and failed, so the caller never reached rpc.DialOptions.
func TestIsURLAvailable_ProxiedTargetUnreachableDirectly(t *testing.T) {
	withProxy(t, "http://"+listening(t))
	require.True(t,
		IsURLAvailable(context.Background(), "http://erpc.internal.invalid:4000/main", time.Second),
		"with a reachable proxy configured, availability should follow the proxy")
}

func TestIsURLAvailable_ProxyUnreachable(t *testing.T) {
	withProxy(t, "http://"+closedPort(t))
	require.False(t,
		IsURLAvailable(context.Background(), "http://"+listening(t), time.Second),
		"an unreachable proxy means the client cannot get out, target liveness is irrelevant")
}

func TestIsURLAvailable_UnknownSchemeFailsOpen(t *testing.T) {
	withProxy(t, "")
	require.True(t,
		IsURLAvailable(context.Background(), "unix:///var/run/thing.sock", time.Second),
		"schemes with no well-known port should fail open, as before")
}

func TestHostPort(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"http://h:4000/p", "h:4000"},
		{"http://h/p", "h:80"},
		{"ws://h", "h:80"},
		{"https://h", "h:443"},
		{"wss://h", "h:443"},
		{"unix:///s.sock", ""},
	} {
		u, err := url.Parse(tc.in)
		require.NoError(t, err)
		require.Equal(t, tc.want, hostPort(u), tc.in)
	}
}
