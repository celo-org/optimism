package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const defaultConnectTimeout = 5 * time.Second

func TestIsURLAvailableLocal(t *testing.T) {
	listener, err := net.Listen("tcp4", ":0")
	require.NoError(t, err)
	defer listener.Close()

	a := listener.Addr().String()
	parts := strings.Split(a, ":")
	addr := fmt.Sprintf("http://localhost:%s", parts[1])

	// True & False with ports
	require.True(t, IsURLAvailable(context.Background(), addr, defaultConnectTimeout))
	require.False(t, IsURLAvailable(context.Background(), "http://localhost:0", defaultConnectTimeout))

	// Fail open if we don't recognize the scheme
	require.True(t, IsURLAvailable(context.Background(), "mailto://example.com", defaultConnectTimeout))

}

func TestIsURLAvailableNonLocal(t *testing.T) {
	// Probe the targets directly: through a proxy the fake domains look reachable.
	swapProxyResolver(t, noProxy)

	if !IsURLAvailable(context.Background(), "http://example.com", defaultConnectTimeout) {
		t.Skip("No internet connection found, skipping this test")
	}

	// True without ports. http & https
	require.True(t, IsURLAvailable(context.Background(), "http://example.com", defaultConnectTimeout))
	require.True(t, IsURLAvailable(context.Background(), "http://example.com/hello", defaultConnectTimeout))
	require.True(t, IsURLAvailable(context.Background(), "https://example.com", defaultConnectTimeout))
	require.True(t, IsURLAvailable(context.Background(), "https://example.com/hello", defaultConnectTimeout))

	// True without ports. ws & wss
	require.True(t, IsURLAvailable(context.Background(), "ws://example.com", defaultConnectTimeout))
	require.True(t, IsURLAvailable(context.Background(), "ws://example.com/hello", defaultConnectTimeout))
	require.True(t, IsURLAvailable(context.Background(), "wss://example.com", defaultConnectTimeout))
	require.True(t, IsURLAvailable(context.Background(), "wss://example.com/hello", defaultConnectTimeout))

	// False without ports
	require.False(t, IsURLAvailable(context.Background(), "http://fakedomainnamethatdoesnotexistandshouldneverexist.com", defaultConnectTimeout))
	require.False(t, IsURLAvailable(context.Background(), "http://fakedomainnamethatdoesnotexistandshouldneverexist.com/hello", defaultConnectTimeout))
	require.False(t, IsURLAvailable(context.Background(), "https://fakedomainnamethatdoesnotexistandshouldneverexist.com", defaultConnectTimeout))
	require.False(t, IsURLAvailable(context.Background(), "https://fakedomainnamethatdoesnotexistandshouldneverexist.com/hello", defaultConnectTimeout))
	require.False(t, IsURLAvailable(context.Background(), "ws://fakedomainnamethatdoesnotexistandshouldneverexist.com", defaultConnectTimeout))
	require.False(t, IsURLAvailable(context.Background(), "ws://fakedomainnamethatdoesnotexistandshouldneverexist.com/hello", defaultConnectTimeout))
	require.False(t, IsURLAvailable(context.Background(), "wss://fakedomainnamethatdoesnotexistandshouldneverexist.com", defaultConnectTimeout))
	require.False(t, IsURLAvailable(context.Background(), "wss://fakedomainnamethatdoesnotexistandshouldneverexist.com/hello", defaultConnectTimeout))
}

// TestIsURLAvailableProxy checks that the probe dials the proxy, not the
// target, when one is configured. The proxy is a live listener and the target
// is a closed port, so a probe that ignores the proxy fails.
func TestIsURLAvailableProxy(t *testing.T) {
	proxy, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	defer proxy.Close()
	proxyAddr := proxy.Addr().String()

	// A port that was just released refuses connections immediately.
	closed, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	unreachable := closed.Addr().String()
	require.NoError(t, closed.Close())

	// proxyAt acts like http.ProxyFromEnvironment with HTTP_PROXY and
	// HTTPS_PROXY set to addr: it returns the proxy for http and https URLs
	// and nil for anything else. It also records the URL it was asked about.
	// The real resolver is not used because it never proxies localhost, and
	// this test runs everything on localhost.
	proxyAt := func(addr string, asked **url.URL) func(*http.Request) (*url.URL, error) {
		return func(req *http.Request) (*url.URL, error) {
			*asked = req.URL
			switch req.URL.Scheme {
			case "http", "https":
				return &url.URL{Scheme: "http", Host: addr}, nil
			default:
				return nil, nil
			}
		}
	}

	for _, tc := range []struct {
		scheme string
		// Scheme the proxy resolver must be asked about.
		lookupScheme string
	}{
		{scheme: "http", lookupScheme: "http"},
		{scheme: "https", lookupScheme: "https"},
		{scheme: "ws", lookupScheme: "http"},
		{scheme: "wss", lookupScheme: "https"},
	} {
		t.Run(tc.scheme, func(t *testing.T) {
			target := tc.scheme + "://" + unreachable + "/rpc"

			// No proxy: the probe reaches the closed target and fails.
			swapProxyResolver(t, noProxy)
			require.False(t, IsURLAvailable(context.Background(), target, defaultConnectTimeout))

			// Proxy configured: only the proxy is dialed, so the probe succeeds.
			var asked *url.URL
			swapProxyResolver(t, proxyAt(proxyAddr, &asked))
			require.True(t, IsURLAvailable(context.Background(), target, defaultConnectTimeout))
			require.NotNil(t, asked, "proxy resolver was not consulted")
			require.Equal(t, tc.lookupScheme, asked.Scheme)
			require.Equal(t, unreachable, asked.Host, "resolver must see the original host")

			// Dead proxy: the probe fails even though the target is reachable.
			swapProxyResolver(t, proxyAt(unreachable, &asked))
			reachable := tc.scheme + "://" + proxyAddr + "/rpc"
			require.False(t, IsURLAvailable(context.Background(), reachable, defaultConnectTimeout))
		})
	}
}

// noProxy is a proxy resolver that never returns a proxy.
func noProxy(*http.Request) (*url.URL, error) { return nil, nil }

// swapProxyResolver replaces proxyForRequest until the test ends. Tests using
// it must not run in parallel.
func swapProxyResolver(t *testing.T, fn func(*http.Request) (*url.URL, error)) {
	t.Helper()
	prev := proxyForRequest
	proxyForRequest = fn
	t.Cleanup(func() { proxyForRequest = prev })
}
