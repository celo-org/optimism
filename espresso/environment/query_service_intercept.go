package environment

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"time"

	espressoTaggedBase64 "github.com/EspressoSystems/espresso-network/sdks/go/tagged-base64"
	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/coder/websocket"
	"github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
)

// InterceptHandleDecision is an enum that represents a decision on how to
// handle the http handler request for the specific request.
//
// It is meant to represent the specific behavior that the user would like to
// happen to a given request, without needing to worry about the implementation
// for how to make that behavior happen.
type InterceptHandleDecision int

const (
	// DecisionProxy means that the request should be proxied unmodified to the
	// target service (the Espresso Dev Node).
	DecisionProxy InterceptHandleDecision = iota

	// DecisionReportSubmitSuccessWhileDropped means that the request should
	// be handled by simulating a successful transaction submission, but
	// without actually submitting the transaction to the Espresso Dev Node.
	DecisionReportSubmitSuccessWhileDropped

	// DecisionReportServerUnreachable means that the request should be
	// handled by returning an error indicating that the Espresso Dev Node
	// was unreachable to the client.
	DecisionReportServerUnreachable
)

// builderHandler is a method that will build the appropriate HTTP handler based
// on the provided decision.
func (d InterceptHandleDecision) buildHandler(client *http.Client, baseURL url.URL) http.Handler {
	switch d {
	case DecisionProxy:
		return &proxyRequest{
			client:  client,
			baseURL: baseURL,
		}

	case DecisionReportSubmitSuccessWhileDropped:
		return fakeSubmitTransactionSuccess{}

	case DecisionReportServerUnreachable:
		return reportServerUnreachable{}

	default:
		return nil
	}
}

// InterceptHandlerDecider is an interface that defines a method for
// deciding how it should handle a given HTTP request.
//
// The idea is to make it simple for the user to implement their own logic for
// how to determine how to handle a request without needing to worry about the
// implementation details of the proxying, or the specific handling cases of
// his / her desired behaviors.
type InterceptHandlerDecider interface {
	DecideHowToHandleRequest(w http.ResponseWriter, r *http.Request) InterceptHandleDecision
}

// defaultInterceptHandlerDecider is a simple implementation of the
// InterceptHandlerDecider interface that always returns a proxy decision.
type defaultInterceptHandlerDecider struct{}

// DecideHowToHandleRequest implements InterceptHandlerDecider
func (defaultInterceptHandlerDecider) DecideHowToHandleRequest(w http.ResponseWriter, r *http.Request) InterceptHandleDecision {
	return DecisionProxy
}

// proxyRequest is a simple HTTP handler that proxies requests to the given
// baseURL, utilizing the given http.Client.
type proxyRequest struct {
	client  *http.Client
	baseURL url.URL
}

// ServeHTTP implements http.Handler
func (p *proxyRequest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if this is a websocket stream request
	if isWebSocketStreamRequest(r) {
		p.proxyWebSocket(w, r)
		return
	}

	// Handle regular HTTP requests
	defer r.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest(r.Method, p.baseURL.JoinPath(r.URL.Path).String(), buf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy over the headers
	for k, v := range r.Header {
		req.Header.Set(k, v[0])
	}

	res, err := p.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	buf.Reset()
	if _, err := io.Copy(buf, res.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(res.StatusCode)
	for k, v := range res.Header {
		w.Header().Set(k, v[0])
	}

	// Write the proxy response contents
	if _, err := io.Copy(w, buf); err != nil {
		// If we encounter an error here, it will be difficult to actually
		// handle it at this point, as we've already sent the response headers.
		//
		// The best we can do at this point, is log the error.
		_ = err
		return
	}
}

// proxyWebSocket handles websocket upgrade and proxying
func (p *proxyRequest) proxyWebSocket(w http.ResponseWriter, r *http.Request) {
	// Accept the websocket connection from the client
	clientConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close(websocket.StatusInternalError, "proxy error")

	// Create websocket URL for the backend
	backendURL := p.baseURL
	if backendURL.Scheme == "https" {
		backendURL.Scheme = "wss"
	} else {
		backendURL.Scheme = "ws"
	}
	backendURL.Path = r.URL.Path
	backendURL.RawQuery = r.URL.RawQuery

	// Connect to the backend websocket
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//nolint:bodyclose // Not applicable to coder/websocket. From the websocket.Dial docs: "You never need to close resp.Body yourself."
	backendConn, _, err := websocket.Dial(ctx, backendURL.String(), &websocket.DialOptions{})
	if err != nil {
		clientConn.Close(websocket.StatusInternalError, "backend connection failed")
		return
	}
	defer backendConn.Close(websocket.StatusNormalClosure, "")

	// Proxy messages bidirectionally
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// Client to backend
	go func() {
		defer cancel()
		for {
			msgType, data, err := clientConn.Read(ctx)
			if err != nil {
				return
			}
			err = backendConn.Write(ctx, msgType, data)
			if err != nil {
				return
			}
		}
	}()

	// Backend to client
	go func() {
		defer cancel()
		for {
			msgType, data, err := backendConn.Read(ctx)
			if err != nil {
				return
			}
			err = clientConn.Write(ctx, msgType, data)
			if err != nil {
				return
			}
		}
	}()

	// Wait until context is cancelled (one of the goroutines finished)
	<-ctx.Done()

	// Close connections gracefully
	clientConn.Close(websocket.StatusNormalClosure, "")
	backendConn.Close(websocket.StatusNormalClosure, "")
}

// fakeSubmitTransactionSuccess is a simple HTTP handler that simulates a
// successful transaction submission by returning a fake commit hash.
type fakeSubmitTransactionSuccess struct{}

// generateCommitForSubmitTransaction generates a commit hash for the
// transaction in the request body. This is a fake implementation that
// simulates a successful transaction submission by returning a commit hash
// that won't collide with the real transaction commit hashes.
func generateCommitForSubmitTransaction(r *http.Request) (*espressoTaggedBase64.TaggedBase64, error) {
	defer r.Body.Close()

	var txn espressoCommon.Transaction
	if err := json.NewDecoder(r.Body).Decode(&txn); err != nil {
		// Unable to decode, this is a problem?
		var emptyHash [32]byte
		return espressoTaggedBase64.New("FAKE", emptyHash[:])
	}

	commit := txn.Commit()
	return espressoTaggedBase64.New("FAKE", commit[:])
}

// ServeHTTP implements http.Handler
func (fakeSubmitTransactionSuccess) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// We could do a lot of effort to validate the request, and return a
	// hash that is actually representative of the transaction that was
	// just submitted. In some cases we may actually want this sort of
	// validated behavior, but it's very simple to just return any hash
	// instead.

	// We should probably validate the request contents and format here, but
	// we will just assume the settings.
	defer r.Body.Close()
	hash, err := generateCommitForSubmitTransaction(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contents, err := json.Marshal(hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(contents)
}

// reportServerUnreachable is a simple HTTP handler that simulates a load
// balancer, or some other intermediary, returning an error indicating that the
// target handling service is unreachable.
type reportServerUnreachable struct{}

// ServeHTTP implements http.Handler
func (reportServerUnreachable) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Don't forget to close the request body, though we won't actually read
	// anything from it.
	defer r.Body.Close()

	http.Error(w, "service unreachable", http.StatusServiceUnavailable)
}

// EspressoDevNodeIntercept is a struct that is a Proxy to the Espresso Dev Node.
// It is used to intercept request to the Espresso Dev Node and make decisions
// about handling the requests.  This is useful for simulating failures or
// bad behaviors for Espresso.
type EspressoDevNodeIntercept struct {
	u       url.URL
	client  *http.Client
	decider InterceptHandlerDecider
}

type Rng interface {
	Intn(n int) int
}

// randomRollFakeSubmitTransactionSuccess is a InterceptHandlerDecider that aids
// in the simulation of various transaction submission failures by randomly
// deciding whether to return a successful submission response, to return that
// the service is unavailable or to proxy the request to the Espresso Dev Node.
type randomRollFakeSubmitTransactionSuccess struct {
	// n the upper end of the range to roll against.
	n int

	// The fakeSuccessThreshold, under which will trigger a faked submission
	// success.
	fakeSuccessThreshold int

	// fakeServiceUnavailableThreshold is the threshold under which the
	// decision will return a simulated service unavailable error.
	fakeServiceUnavailableThreshold int

	// the Rng to use to determine the random roll
	r Rng
}

// NewRandomRollFakeSubmitTransactionSuccess creates a new
// InterceptHandlerDecider that will proxy all requests to the espresso dev
// node except for the submit transaction requests.
//
// When a submit transaction request is received it will use the provided Rng
// roll a number between 0 and n-1.
// Depending on the value rolled it will determine the resulting behavior as
// follows:
//   - If the number rolled is less than or equal to the given
//     fakeSuccessThreshold, it will return a simulated successful transaction
//     submission response, while dropping the request ensuring it does not
//     actually reach the Espresso Dev Node.
//   - If the number rolled is less than or equal to the given
//     fakeServiceUnavailableThreshold, it will return a simulated service
//     unavailable error response.
//   - Otherwise, it will proxy the request to the Espresso Dev Node.
//
// NOTE: We only roll once per request, so the thresholds are not cumulative,
// This means if they overlap, then one of the behaviors will never be
// triggered.  However, you can utilize this to your advantage by setting
// the thresholds so that they overlap directly, ensuring you only test
// one of the behaviors.
//
// NOTE: Setting the `fakeSuccessThreshold` value less than `0` will ensure
// that the fake success threshold case is never triggered.
//
// NOTE: Setting the `fakeServiceUnavailableThreshold` value less than or
// equal to `fakeSuccessThreshold` will ensure that the service unavailable
// threshold case is never triggered.
//
// The thresholds are evaluated in order of `fakeSuccessThreshold`, then
// `fakeServiceUnavailableThreshold`, then the default proxy behavior.
// So if you want to ensure that all cases are tested you should specify your
// values with the following constraints:
//   - `fakeSuccessThreshold` >= 0
//   - `fakeServiceUnavailableThreshold` > `fakeSuccessThreshold`
//   - `fakeServiceUnavailableThreshold` < `n`
func NewRandomRollFakeSubmitTransactionSuccess(
	rollUpperRange,
	fakeSuccessThreshold,
	fakeServiceUnavailableThreshold int,
	r Rng,
) InterceptHandlerDecider {
	return &randomRollFakeSubmitTransactionSuccess{
		n:                               rollUpperRange,
		fakeSuccessThreshold:            fakeSuccessThreshold,
		fakeServiceUnavailableThreshold: fakeServiceUnavailableThreshold,
		r:                               r,
	}
}

// requestMatchesPath checks if the HTTP request matches the specified method
func requestMatchesPath(r *http.Request, method string, pathMatcher func(string) bool) bool {
	return r.Method == method && r.URL != nil && pathMatcher(r.URL.Path)
}

// stringEquals is a helper function that returns a function that checks if
// a given path string equals the specified string.
func stringEquals(s string) func(string) bool {
	return func(path string) bool {
		return path == s
	}
}

// isSubmitTransactionRequest represents the different variations of the submit
// transaction endpoint that we can utilize or support.
func isSubmitTransactionRequest(r *http.Request) bool {
	return requestMatchesPath(r, http.MethodPost, stringEquals("/submit/submit")) ||
		requestMatchesPath(r, http.MethodPost, stringEquals("/v0/submit/submit"))
}

// isWebSocketStreamRequest checks if the request is a websocket request
// matching the pattern "vN/stream/*" where N is an integer.
func isWebSocketStreamRequest(r *http.Request) bool {
	matched, _ := regexp.MatchString(`/stream/`, r.URL.Path)
	return matched
}

// DecideHowToHandleRequest implements InterceptHandlerDecider
func (d *randomRollFakeSubmitTransactionSuccess) DecideHowToHandleRequest(w http.ResponseWriter, r *http.Request) InterceptHandleDecision {
	if isSubmitTransactionRequest(r) {
		// We want to randomly simulate a failure in the transaction
		// submission.  We compare our random roll against our thresholds in
		// order to return the appropriate decision for how to handle the
		// request.
		roll := d.r.Intn(d.n)
		if roll <= d.fakeSuccessThreshold {
			return DecisionReportSubmitSuccessWhileDropped
		} else if roll <= d.fakeServiceUnavailableThreshold {
			return DecisionReportServerUnreachable
		}
	}
	return DecisionProxy
}

// ServerHTTP implements http.Handler
func (e *EspressoDevNodeIntercept) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decision := e.decider.DecideHowToHandleRequest(w, r)
	handler := decision.buildHandler(e.client, e.u)
	handler.ServeHTTP(w, r)
}

// createEspressoProxyOption will return a Batch CLIConfig option that will
// replace the Espresso URL with the URL of the proxy server.
func createEspressoProxyOption(ctx *E2eDevnetLauncherContext, proxy *EspressoDevNodeIntercept, server *httptest.Server) func(*batcher.CLIConfig, *e2esys.System) {
	return func(cfg *batcher.CLIConfig, sys *e2esys.System) {
		if ctx.Error != nil {
			return
		}

		if len(cfg.Espresso.QueryServiceURLs) == 0 {
			// This should be being called after the Espresso
			// Dev Node is Already Live.
			// Without an Espresso URL, we cannot proceed.
			return
		}

		u, err := url.Parse(cfg.Espresso.QueryServiceURLs[0])
		if err != nil || u == nil {
			// We encountered an error
			ctx.Error = err
			return
		}

		// Set the proxy
		proxy.u = *u
		// Replace the Espresso URL with the proxy URL
		// We need to provide at least 2 URLs to create a valid multiple nodes client
		cfg.Espresso.QueryServiceURLs = []string{server.URL, server.URL}
	}
}

// EspressoDevNodeInterceptOption is a function that modifies the
// EspressoDevNodeIntercept configuration.
type EspressoDevNodeInterceptOption func(*EspressoDevNodeIntercept)

// SetDecider sets the InterceptHandlerDecider for the EspressoDevNodeIntercept.
func SetDecider(decider InterceptHandlerDecider) EspressoDevNodeInterceptOption {
	return func(e *EspressoDevNodeIntercept) {
		e.decider = decider
	}
}

// SetHTTPClient sets the HTTP client for the EspressoDevNodeIntercept.
func SetHTTPClient(client *http.Client) EspressoDevNodeInterceptOption {
	return func(e *EspressoDevNodeIntercept) {
		e.client = client
	}
}

// SetupQueryServiceIntercept sets up an intercept traffic headed for the
// Query Service for the Espresso Dev Node
func SetupQueryServiceIntercept(options ...EspressoDevNodeInterceptOption) (*EspressoDevNodeIntercept, *httptest.Server, E2eDevnetLauncherOption) {
	// Start a Server to proxy requests to Espresso
	proxy := &EspressoDevNodeIntercept{
		client:  http.DefaultClient,
		decider: defaultInterceptHandlerDecider{},
	}

	for _, opt := range options {
		opt(proxy)
	}

	// Start up a local http server to handle the requests
	server := httptest.NewServer(proxy)

	return proxy, server, func(ctx *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				{
					Key:        "espresso-proxy",
					BatcherMod: createEspressoProxyOption(ctx, proxy, server),
				},
			},
		}
	}
}
