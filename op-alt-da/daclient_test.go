package altda

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/stretchr/testify/require"
)

func TestDAClientPrecomputed(t *testing.T) {
	store := NewMemStore()
	logger := testlog.Logger(t, log.LevelDebug)

	ctx := context.Background()

	server := NewDAServer("127.0.0.1", 0, store, logger, false)

	require.NoError(t, server.Start())

	cfg := CLIConfig{
		Enabled:      true,
		DAServerURL:  server.HttpEndpoint(),
		VerifyOnRead: true,
	}
	require.NoError(t, cfg.Check())

	client, err := cfg.NewDAClient()
	require.NoError(t, err)

	rng := rand.New(rand.NewSource(1234))

	input := RandomData(rng, 2000)

	comm, err := client.SetInput(ctx, input)
	require.NoError(t, err)

	require.Equal(t, comm, NewKeccak256Commitment(input))

	stored, err := client.GetInput(ctx, comm, 0)
	require.NoError(t, err)

	require.Equal(t, input, stored)

	// set a bad commitment in the store
	require.NoError(t, store.Put(ctx, comm.Encode(), []byte("bad data")))

	_, err = client.GetInput(ctx, comm, 0)
	require.ErrorIs(t, err, ErrCommitmentMismatch)

	// test not found error
	comm = NewKeccak256Commitment(RandomData(rng, 32))
	_, err = client.GetInput(ctx, comm, 0)
	require.ErrorIs(t, err, ErrNotFound)

	// test storing bad data
	_, err = client.SetInput(ctx, []byte{})
	require.ErrorIs(t, err, ErrInvalidInput)

	// server not responsive
	require.NoError(t, server.Stop())
	_, err = client.SetInput(ctx, input)
	require.Error(t, err)

	_, err = client.GetInput(ctx, NewKeccak256Commitment(input), 0)
	require.Error(t, err)
}

func TestDAClientService(t *testing.T) {
	store := NewMemStore()
	logger := testlog.Logger(t, log.LevelDebug)

	ctx := context.Background()

	server := NewDAServer("127.0.0.1", 0, store, logger, false)

	require.NoError(t, server.Start())

	cfg := CLIConfig{
		Enabled:      true,
		DAServerURL:  server.HttpEndpoint(),
		VerifyOnRead: false,
		GenericDA:    false,
	}
	require.NoError(t, cfg.Check())

	client, err := cfg.NewDAClient()
	require.NoError(t, err)

	rng := rand.New(rand.NewSource(1234))

	input := RandomData(rng, 2000)

	comm, err := client.SetInput(ctx, input)
	require.NoError(t, err)

	require.Equal(t, comm.String(), NewKeccak256Commitment(input).String())

	stored, err := client.GetInput(ctx, comm, 0)
	require.NoError(t, err)

	require.Equal(t, input, stored)

	// set a bad commitment in the store
	require.NoError(t, store.Put(ctx, comm.Encode(), []byte("bad data")))

	// assert no error as generic commitments cannot be verified client side
	_, err = client.GetInput(ctx, comm, 0)
	require.NoError(t, err)

	// test not found error
	comm = NewKeccak256Commitment(RandomData(rng, 32))
	_, err = client.GetInput(ctx, comm, 0)
	require.ErrorIs(t, err, ErrNotFound)

	// test storing bad data
	_, err = client.SetInput(ctx, []byte{})
	require.ErrorIs(t, err, ErrInvalidInput)

	// server not responsive
	require.NoError(t, server.Stop())
	_, err = client.SetInput(ctx, input)
	require.Error(t, err)

	_, err = client.GetInput(ctx, NewKeccak256Commitment(input), 0)
	require.Error(t, err)
}

// TestDAClientGetInputTeapot checks that a 418 TEAPOT response only yields a DropEigenDACommitmentError,
// the error that makes derivation drop the commitment, when its body carries a status code between 1 and 4.
// Any other body must yield an ordinary error, which derivation retries.
func TestDAClientGetInputTeapot(t *testing.T) {
	tests := []struct {
		name string
		body string
		// expectedDrop is the DropEigenDACommitmentError expected from GetInput, or nil if an ordinary error is expected.
		expectedDrop *DropEigenDACommitmentError
	}{
		{name: "status code 1", body: `{"StatusCode": 1, "Msg": "invalid cert"}`, expectedDrop: &DropEigenDACommitmentError{StatusCode: 1, Msg: "invalid cert"}},
		{name: "status code 2", body: `{"StatusCode": 2, "Msg": "invalid cert"}`, expectedDrop: &DropEigenDACommitmentError{StatusCode: 2, Msg: "invalid cert"}},
		{name: "status code 3", body: `{"StatusCode": 3, "Msg": "invalid cert"}`, expectedDrop: &DropEigenDACommitmentError{StatusCode: 3, Msg: "invalid cert"}},
		{name: "status code 4 without msg", body: `{"StatusCode": 4}`, expectedDrop: &DropEigenDACommitmentError{StatusCode: 4}},
		{name: "status code missing", body: `{"Msg": "invalid cert"}`},
		{name: "empty object", body: `{}`},
		{name: "status code 0", body: `{"StatusCode": 0, "Msg": "invalid cert"}`},
		{name: "status code 5", body: `{"StatusCode": 5, "Msg": "invalid cert"}`},
		{name: "status code negative", body: `{"StatusCode": -1, "Msg": "invalid cert"}`},
		{name: "body is not json", body: `I'm a teapot`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			client := NewDAClient(server.URL, false, false)
			_, err := client.GetInput(context.Background(), NewGenericCommitment([]byte("cert")), 0)
			require.Error(t, err)

			var dropErr DropEigenDACommitmentError
			if tc.expectedDrop == nil {
				require.False(t, errors.As(err, &dropErr), "must not drop the commitment, got: %v", err)
				return
			}
			require.ErrorAs(t, err, &dropErr)
			require.Equal(t, *tc.expectedDrop, dropErr)
		})
	}
}

// TestDAClientGetInputClosesErrorBody checks that GetInput releases the connection of a response
// it rejects by its status code, even though it does not read that response's body.
func TestDAClientGetInputClosesErrorBody(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var mu sync.Mutex
			openConns := make(map[net.Conn]struct{})

			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte("error details that the client does not read"))
			}))
			server.Config.ConnState = func(conn net.Conn, state http.ConnState) {
				mu.Lock()
				defer mu.Unlock()
				if state == http.StateClosed {
					delete(openConns, conn)
				} else {
					openConns[conn] = struct{}{}
				}
			}
			server.Start()
			defer server.Close()

			client := NewDAClient(server.URL, false, false)
			_, err := client.GetInput(context.Background(), NewGenericCommitment([]byte("cert")), 0)
			require.Error(t, err)

			require.Eventually(t, func() bool {
				mu.Lock()
				defer mu.Unlock()
				return len(openConns) == 0
			}, 10*time.Second, 10*time.Millisecond, "connection of the rejected response was left open")
		})
	}
}
