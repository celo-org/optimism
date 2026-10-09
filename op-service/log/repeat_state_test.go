package log

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/log"
)

// safeTestRecorder is a thread-safe slog.Handler that captures records for
// assertions in concurrent tests.
type safeTestRecorder struct {
	mu      sync.Mutex
	records []slog.Record
}

func (r *safeTestRecorder) Enabled(context.Context, slog.Level) bool { return true }

func (r *safeTestRecorder) Handle(_ context.Context, rec slog.Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, rec)
	return nil
}

func (r *safeTestRecorder) WithAttrs([]slog.Attr) slog.Handler { return r }
func (r *safeTestRecorder) WithGroup(string) slog.Handler      { return r }

func (r *safeTestRecorder) GetRecords() []slog.Record {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]slog.Record, len(r.records))
	copy(result, r.records)
	return result
}

func (r *safeTestRecorder) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.records)
}

// fakeClock returns the time stored at *t, allowing the test to advance time
// deterministically without sleeping.
func fakeClock(now *time.Time) func() time.Time {
	return func() time.Time { return *now }
}

// matchingRecords returns records whose level and message match the given
// level and msg. Pass an empty msg to match any message.
func matchingRecords(records []slog.Record, level slog.Level, msg string) []slog.Record {
	var out []slog.Record
	for _, r := range records {
		if r.Level != level {
			continue
		}
		if msg != "" && r.Message != msg {
			continue
		}
		out = append(out, r)
	}
	return out
}

// attrValue extracts the value of the named attribute, or nil if absent.
func attrValue(r slog.Record, key string) any {
	var v any
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == key {
			v = a.Value.Any()
			return false
		}
		return true
	})
	return v
}

func newCapturingLogger() (log.Logger, *safeTestRecorder) {
	rec := new(safeTestRecorder)
	return log.NewLogger(rec), rec
}

func TestRepeatStateLogger_FirstWarnEmitsThenSuppresses(t *testing.T) {
	lgr, rec := newCapturingLogger()
	now := time.Unix(0, 0)
	r := NewRepeatStateLogger()
	r.clock = fakeClock(&now)

	r.Warn(lgr, "k1", "degraded", "err", "boom")
	now = now.Add(1 * time.Second)
	r.Warn(lgr, "k1", "degraded", "err", "boom")
	now = now.Add(1 * time.Second)
	r.Warn(lgr, "k1", "degraded", "err", "boom")

	warns := matchingRecords(rec.GetRecords(), slog.LevelWarn, "degraded")
	require.Len(t, warns, 1, "only the first observation should emit a log")

	require.Equal(t, "boom", attrValue(warns[0], "err"))
	require.Nil(t, attrValue(warns[0], "occurrences"), "first emission should not carry an occurrences count")
}

func TestRepeatStateLogger_ReminderAfterInterval(t *testing.T) {
	lgr, rec := newCapturingLogger()
	now := time.Unix(0, 0)
	r := NewRepeatStateLogger()
	r.clock = fakeClock(&now)

	// Initial emission.
	r.Warn(lgr, "k1", "degraded")

	// 9 silent observations, every 30s. Cumulative 4m30s, still under the 5m threshold.
	for i := 0; i < 9; i++ {
		now = now.Add(30 * time.Second)
		r.Warn(lgr, "k1", "degraded")
	}
	warns := matchingRecords(rec.GetRecords(), slog.LevelWarn, "degraded")
	require.Len(t, warns, 1, "no reminder before the interval has elapsed")

	// Cross the threshold.
	now = now.Add(31 * time.Second)
	r.Warn(lgr, "k1", "degraded")

	warns = matchingRecords(rec.GetRecords(), slog.LevelWarn, "degraded")
	require.Len(t, warns, 2, "reminder should fire once the interval has elapsed")

	// Reminder reports cumulative occurrences: 1 initial + 9 silent + 1 reminder = 11.
	require.EqualValues(t, int64(11), attrValue(warns[1], "occurrences"))
	// Duration since firstSeen is 9*30s + 31s = 5m1s, rounded to nearest second.
	require.Equal(t, 5*time.Minute+1*time.Second, attrValue(warns[1], "duration"))
}

func TestRepeatStateLogger_KeysAreIndependent(t *testing.T) {
	lgr, rec := newCapturingLogger()
	now := time.Unix(0, 0)
	r := NewRepeatStateLogger()
	r.clock = fakeClock(&now)

	r.Warn(lgr, "k1", "first state")
	r.Warn(lgr, "k2", "second state")
	// Both keys are now active. Repeats for either should be suppressed.
	r.Warn(lgr, "k1", "first state")
	r.Warn(lgr, "k2", "second state")

	require.Len(t, matchingRecords(rec.GetRecords(), slog.LevelWarn, "first state"), 1)
	require.Len(t, matchingRecords(rec.GetRecords(), slog.LevelWarn, "second state"), 1)

	// Clearing one key must not affect the other.
	r.Clear(lgr, "k1", "first recovered")
	r.Warn(lgr, "k2", "second state") // still suppressed
	require.Len(t, matchingRecords(rec.GetRecords(), slog.LevelWarn, "second state"), 1)
}

func TestRepeatStateLogger_ClearEmitsRecoveryWhenActive(t *testing.T) {
	lgr, rec := newCapturingLogger()
	now := time.Unix(0, 0)
	r := NewRepeatStateLogger()
	r.clock = fakeClock(&now)

	r.Warn(lgr, "k1", "degraded")
	now = now.Add(2 * time.Second)
	r.Warn(lgr, "k1", "degraded") // suppressed, but increments totalOccurrences
	now = now.Add(3 * time.Second)
	r.Clear(lgr, "k1", "recovered", "extra", "ctx")

	infos := matchingRecords(rec.GetRecords(), slog.LevelInfo, "recovered")
	require.Len(t, infos, 1)
	require.Equal(t, 5*time.Second, attrValue(infos[0], "duration"))
	require.EqualValues(t, int64(2), attrValue(infos[0], "occurrences"))
	require.Equal(t, "ctx", attrValue(infos[0], "extra"))
}

func TestRepeatStateLogger_ClearWhenInactiveIsNoop(t *testing.T) {
	lgr, rec := newCapturingLogger()
	now := time.Unix(0, 0)
	r := NewRepeatStateLogger()
	r.clock = fakeClock(&now)

	r.Clear(lgr, "k1", "recovered")

	require.Empty(t, matchingRecords(rec.GetRecords(), slog.LevelInfo, "recovered"))
}

func TestRepeatStateLogger_FreshAfterClear(t *testing.T) {
	lgr, rec := newCapturingLogger()
	now := time.Unix(0, 0)
	r := NewRepeatStateLogger()
	r.clock = fakeClock(&now)

	r.Warn(lgr, "k1", "degraded")
	r.Warn(lgr, "k1", "degraded") // suppressed
	r.Clear(lgr, "k1", "recovered")

	// After Clear, the next Warn should emit again as a fresh first observation.
	r.Warn(lgr, "k1", "degraded")

	warns := matchingRecords(rec.GetRecords(), slog.LevelWarn, "degraded")
	require.Len(t, warns, 2, "first Warn after Clear should emit")
	require.Nil(t, attrValue(warns[1], "occurrences"), "fresh emission should not carry an occurrences count")
}

func TestRepeatStateLogger_ConcurrentCallersDoNotRace(t *testing.T) {
	lgr, _ := newCapturingLogger()
	r := NewRepeatStateLogger()

	const goroutines = 32
	const callsPerG = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)
	var clears atomic.Int64
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			key := "k" + string(rune('a'+(id%4)))
			for i := 0; i < callsPerG; i++ {
				r.Warn(lgr, key, "degraded", "g", id)
				if i%50 == 0 {
					r.Clear(lgr, key, "recovered")
					clears.Add(1)
				}
			}
		}(g)
	}
	wg.Wait()
	// No assertion on counts — this test exists to flush out races under -race
	// and to assert the logger does not panic under concurrent Warn/Clear.
	require.Greater(t, clears.Load(), int64(0))
}
