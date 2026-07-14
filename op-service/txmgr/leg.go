package txmgr

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
)

// leg holds everything belonging to one leg of a SendPairAsync pair: the
// signed transaction, the context its submission loop runs on (cancelled by
// the pair when it resolves), and the channel its final response is delivered
// on.
type leg struct {
	sendContext context.Context
	cancelFunc  context.CancelFunc
	tx          *types.Transaction
	responseCh  chan SendResponse
}

// legResult is the outcome of one leg's submission loop.
type legResult struct {
	receipt *types.Receipt
	err     error
}

// send runs the leg's transaction through the manager's full submission loop
// on the leg's own context.
func (l *leg) send(m *SimpleTxManager) legResult {
	receipt, err := m.sendTx(l.sendContext, l.tx)
	return legResult{receipt, err}
}

// respond delivers the leg's final response on its response channel.
func (l *leg) respond(res legResult) {
	l.responseCh <- SendResponse{Receipt: res.receipt, Nonce: l.tx.Nonce(), Err: res.err}
}
