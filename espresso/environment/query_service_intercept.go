package environment

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoTaggedBase64 "github.com/EspressoSystems/espresso-network/sdks/go/tagged-base64"
	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/ethereum-optimism/optimism/espresso"
	"github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
)

// InterceptHandleDecision is an enum that represents a decision on how to
// handle a transaction submission to Espresso.
//
// It is meant to represent the specific behavior that the user would like to
// happen to a given submission, without needing to worry about the
// implementation for how to make that behavior happen.
type InterceptHandleDecision int

const (
	// DecisionProxy means that the submission should be forwarded unmodified
	// to the underlying Espresso client.
	DecisionProxy InterceptHandleDecision = iota

	// DecisionReportSubmitSuccessWhileDropped means that the submission
	// should be reported as successful, without actually being submitted to
	// the underlying Espresso client.
	DecisionReportSubmitSuccessWhileDropped

	// DecisionReportServerUnreachable means that the submission should fail
	// with an error indicating that the Espresso query service was
	// unreachable.
	DecisionReportServerUnreachable
)

// ErrEspressoServiceUnavailable is the error returned for submissions handled
// with DecisionReportServerUnreachable. Like the real client's server errors,
// it wraps espressoClient.ErrEphemeral (retryable).
var ErrEspressoServiceUnavailable = fmt.Errorf("%w: espresso intercept: service unavailable", espressoClient.ErrEphemeral)

// InterceptHandlerDecider decides how a given transaction submission should be
// handled.
//
// The idea is to make it simple for the user to implement their own logic for
// how to handle a submission without needing to worry about the details of
// the interception or the specific handling of their desired behaviors.
type InterceptHandlerDecider interface {
	DecideHowToHandleSubmission(tx espressoCommon.Transaction) InterceptHandleDecision
}

// defaultInterceptHandlerDecider forwards every submission.
type defaultInterceptHandlerDecider struct{}

// DecideHowToHandleSubmission implements InterceptHandlerDecider
func (defaultInterceptHandlerDecider) DecideHowToHandleSubmission(espressoCommon.Transaction) InterceptHandleDecision {
	return DecisionProxy
}

type Rng interface {
	Intn(n int) int
}

// randomRollFakeSubmitTransactionSuccess is an InterceptHandlerDecider that
// aids in the simulation of various transaction submission failures by
// randomly deciding whether to report a successful submission while dropping
// it, to report that the service is unavailable, or to forward the submission.
type randomRollFakeSubmitTransactionSuccess struct {
	// n the upper end of the range to roll against.
	n int

	// The fakeSuccessThreshold, under which will trigger a faked submission
	// success.
	fakeSuccessThreshold int

	// fakeServiceUnavailableThreshold is the threshold under which the
	// decision will return a simulated service unavailable error.
	fakeServiceUnavailableThreshold int

	// mu guards r: the batcher submits from several workers concurrently.
	mu sync.Mutex
	// the Rng to use to determine the random roll
	r Rng
}

// NewRandomRollFakeSubmitTransactionSuccess creates a new
// InterceptHandlerDecider for transaction submissions.
//
// For each submission it will use the provided Rng to roll a number between
// 0 and n-1. Depending on the value rolled it will determine the resulting
// behavior as follows:
//   - If the number rolled is less than or equal to the given
//     fakeSuccessThreshold, it will report a successful submission while
//     dropping the transaction, ensuring it does not actually reach Espresso.
//   - If the number rolled is less than or equal to the given
//     fakeServiceUnavailableThreshold, it will fail the submission with
//     ErrEspressoServiceUnavailable.
//   - Otherwise, it will forward the submission to Espresso.
//
// NOTE: We only roll once per submission, so the thresholds are not
// cumulative. This means if they overlap, then one of the behaviors will never
// be triggered. However, you can utilize this to your advantage by setting
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
// `fakeServiceUnavailableThreshold`, then the default forwarding behavior.
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

// DecideHowToHandleSubmission implements InterceptHandlerDecider
func (d *randomRollFakeSubmitTransactionSuccess) DecideHowToHandleSubmission(espressoCommon.Transaction) InterceptHandleDecision {
	d.mu.Lock()
	roll := d.r.Intn(d.n)
	d.mu.Unlock()

	if roll <= d.fakeSuccessThreshold {
		return DecisionReportSubmitSuccessWhileDropped
	} else if roll <= d.fakeServiceUnavailableThreshold {
		return DecisionReportServerUnreachable
	}
	return DecisionProxy
}

// EspressoClientIntercept wraps the in-memory mock Espresso client that the
// e2e system injects into the batcher, and intercepts the batcher's
// transaction submissions. This is useful for simulating failures or bad
// behaviors for Espresso. All other client calls are forwarded unmodified.
type EspressoClientIntercept struct {
	*espresso.MockEspressoClient

	decider InterceptHandlerDecider

	dropped     atomic.Uint64
	unavailable atomic.Uint64
}

// SubmitTransaction intercepts the submission according to the decider.
func (e *EspressoClientIntercept) SubmitTransaction(ctx context.Context, tx espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
	switch e.decider.DecideHowToHandleSubmission(tx) {
	case DecisionReportSubmitSuccessWhileDropped:
		e.dropped.Add(1)
		return fakeSubmitTransactionCommit(tx)
	case DecisionReportServerUnreachable:
		e.unavailable.Add(1)
		return nil, ErrEspressoServiceUnavailable
	default:
		return e.MockEspressoClient.SubmitTransaction(ctx, tx)
	}
}

// Dropped returns how many submissions were reported successful but dropped.
func (e *EspressoClientIntercept) Dropped() uint64 { return e.dropped.Load() }

// Unavailable returns how many submissions failed with a simulated service
// unavailable error.
func (e *EspressoClientIntercept) Unavailable() uint64 { return e.unavailable.Load() }

// fakeSubmitTransactionCommit returns a hash for a dropped transaction. The
// "FAKE" tag ensures it never collides with a hash the client reports for a
// transaction that was actually submitted.
func fakeSubmitTransactionCommit(tx espressoCommon.Transaction) (*espressoTaggedBase64.TaggedBase64, error) {
	commit := tx.Commit()
	return espressoTaggedBase64.New("FAKE", commit[:])
}

// EspressoClientInterceptOption is a function that modifies the
// EspressoClientIntercept configuration.
type EspressoClientInterceptOption func(*EspressoClientIntercept)

// SetDecider sets the InterceptHandlerDecider for the EspressoClientIntercept.
func SetDecider(decider InterceptHandlerDecider) EspressoClientInterceptOption {
	return func(e *EspressoClientIntercept) {
		e.decider = decider
	}
}

// SetupQueryServiceIntercept sets up an intercept of the Espresso batcher's
// submissions to the Espresso query service. The returned intercept is bound
// to the system's mock Espresso client once the system starts.
func SetupQueryServiceIntercept(options ...EspressoClientInterceptOption) (*EspressoClientIntercept, E2eDevnetLauncherOption) {
	intercept := &EspressoClientIntercept{
		decider: defaultInterceptHandlerDecider{},
	}

	for _, opt := range options {
		opt(intercept)
	}

	return intercept, func(ctx *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				{
					Key: "espresso-intercept",
					BatcherDriverSetupOpts: func(sys *e2esys.System) []batcher.DriverSetupOption {
						if sys.EspressoClient == nil {
							return nil
						}
						intercept.MockEspressoClient = sys.EspressoClient
						return []batcher.DriverSetupOption{batcher.WithEspressoClientOverride(intercept)}
					},
				},
			},
		}
	}
}
