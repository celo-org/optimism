package log

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

// RepeatStateReminderInterval is how often a long-running degraded state
// re-emits a Warn so operators don't lose visibility while the state persists.
const RepeatStateReminderInterval = 5 * time.Minute

// RepeatStateLogger collapses warnings tied to a "degraded state" into a single
// log on entry, periodic reminders while the state persists, and a recovery
// log on exit. This avoids flooding the log debouncer when a tick-driven loop
// fires the same warning every poll interval.
//
// State is keyed by a free-form string supplied by the caller; entries with
// different keys are independent. Safe for concurrent use.
//
// Unlike DebouncingHandler (which operates at the slog.Handler level on a
// short time window), RepeatStateLogger is caller-driven: the caller signals
// state recovery explicitly via Clear, which lets it emit a recovery log a
// handler-level facility can't produce.
type RepeatStateLogger struct {
	mu     sync.Mutex
	states map[string]*repeatStateEntry
	clock  func() time.Time
}

type repeatStateEntry struct {
	firstSeen   time.Time
	lastLogged  time.Time
	occurrences int
}

func NewRepeatStateLogger() *RepeatStateLogger {
	return &RepeatStateLogger{
		states: make(map[string]*repeatStateEntry),
		clock:  time.Now,
	}
}

// Warn reports an observation of a degraded state. The first observation since
// the most recent Clear (or first ever for the key) emits at warn level.
// Subsequent observations within RepeatStateReminderInterval are silently
// counted; once the interval has elapsed a single reminder warn is emitted
// with the cumulative occurrence count and the total duration since the state
// became active.
func (r *RepeatStateLogger) Warn(l log.Logger, key, msg string, ctx ...any) {
	now := r.clock()
	r.mu.Lock()
	e, active := r.states[key]
	if !active {
		r.states[key] = &repeatStateEntry{firstSeen: now, lastLogged: now, occurrences: 1}
		r.mu.Unlock()
		l.Warn(msg, ctx...)
		return
	}
	e.occurrences++
	if now.Sub(e.lastLogged) < RepeatStateReminderInterval {
		r.mu.Unlock()
		return
	}
	occurrences := e.occurrences
	duration := now.Sub(e.firstSeen).Round(time.Second)
	e.lastLogged = now
	r.mu.Unlock()

	l.Warn(msg, append([]any{"occurrences", occurrences, "duration", duration}, ctx...)...)
}

// Clear marks the named state as resolved. If the state was active a single
// info-level recovery log is emitted summarising the duration and total
// occurrences. Calling Clear when the state is not active is a no-op, so it is
// safe to call on every successful tick of the loop.
func (r *RepeatStateLogger) Clear(l log.Logger, key, recoveryMsg string, ctx ...any) {
	r.mu.Lock()
	e, active := r.states[key]
	delete(r.states, key)
	r.mu.Unlock()
	if !active {
		return
	}
	args := append([]any{
		"duration", r.clock().Sub(e.firstSeen).Round(time.Second),
		"occurrences", e.occurrences,
	}, ctx...)
	l.Info(recoveryMsg, args...)
}
