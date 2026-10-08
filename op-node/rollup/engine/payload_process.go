package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

// ErrPayloadInvalid is returned when a payload's execution is invalid.
var ErrPayloadInvalid = errors.New("payload execution invalid")

type PayloadProcessEvent struct {
	// if payload should be promoted to (local) safe (must also be pending safe, see DerivedFrom)
	Concluding bool
	// payload is promoted to pending-safe if non-zero
	DerivedFrom  eth.L1BlockRef
	BuildStarted time.Time

	Envelope *eth.ExecutionPayloadEnvelope
	Ref      eth.L2BlockRef
}

func (ev PayloadProcessEvent) String() string {
	return "payload-process"
}

func (e *EngineController) onPayloadProcess(ctx context.Context, ev PayloadProcessEvent) {
	insertStarted, err := e.processNewPayload(ctx, ev.Envelope, ev.Ref, ev.DerivedFrom)
	if err != nil {
		return
	}
	e.emitter.Emit(ctx, PayloadSuccessEvent{
		Concluding:    ev.Concluding,
		DerivedFrom:   ev.DerivedFrom,
		BuildStarted:  ev.BuildStarted,
		InsertStarted: insertStarted,
		Envelope:      ev.Envelope,
		Ref:           ev.Ref,
	})
}

// processNewPayload performs the NewPayload RPC call.
// It does NOT acquire e.mu (caller is responsible).
// It emits error events for other listeners, but not PayloadSuccessEvent.
func (e *EngineController) processNewPayload(ctx context.Context, envelope *eth.ExecutionPayloadEnvelope, ref eth.L2BlockRef, derivedFrom eth.L1BlockRef) (time.Time, error) {
	rpcCtx, cancel := context.WithTimeout(e.ctx, payloadProcessTimeout)
	defer cancel()

	insertStart := time.Now()
	status, err := e.engine.NewPayload(rpcCtx,
		envelope.ExecutionPayload, envelope.ParentBeaconBlockRoot)
	if err != nil {
		insertErr := fmt.Errorf("failed to insert execution payload: %w", err)
		e.emitter.Emit(ctx, rollup.EngineTemporaryErrorEvent{Err: insertErr})
		return time.Time{}, insertErr
	}
	switch status.Status {
	case eth.ExecutionInvalid, eth.ExecutionInvalidBlockHash:
		// Depending on execution engine, not all block-validity checks run immediately on build-start
		// at the time of the forkchoiceUpdated engine-API call, nor during getPayload.
		if derivedFrom != (eth.L1BlockRef{}) && e.rollupCfg.IsHolocene(derivedFrom.Time) {
			e.emitDepositsOnlyPayloadAttributesRequest(ctx, ref.ParentID(), derivedFrom)
			return time.Time{}, ErrPayloadInvalid
		}

		e.emitter.Emit(ctx, PayloadInvalidEvent{
			Envelope: envelope,
			Err:      eth.NewPayloadErr(envelope.ExecutionPayload, status),
		})
		return time.Time{}, ErrPayloadInvalid
	case eth.ExecutionValid:
		return insertStart, nil
	default:
		statusErr := eth.NewPayloadErr(envelope.ExecutionPayload, status)
		e.emitter.Emit(ctx, rollup.EngineTemporaryErrorEvent{Err: statusErr})
		return time.Time{}, statusErr
	}
}

// ProcessPayload inserts a payload via NewPayload, updates the unsafe head, and finalizes via FCU.
// It emits UnsafeUpdateEvent on success and no PayloadSuccessEvent.
func (e *EngineController) ProcessPayload(ctx context.Context, envelope *eth.ExecutionPayloadEnvelope, ref eth.L2BlockRef, buildStarted time.Time) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if envelope.ExecutionPayload.ParentHash != e.unsafeHead.Hash {
		e.log.Warn("dropping stale sequencer payload, parent is not the unsafe head",
			"payload", ref, "parent", envelope.ExecutionPayload.ParentHash, "unsafe", e.unsafeHead)
		e.requestForkchoiceUpdate(ctx)
		return ErrStaleBuild
	}
	insertStarted, err := e.processNewPayload(ctx, envelope, ref, eth.L1BlockRef{})
	if err != nil {
		return err
	}
	e.finalizePayload(ctx, ref, false, eth.L1BlockRef{}, envelope, buildStarted, insertStarted)
	return nil
}
