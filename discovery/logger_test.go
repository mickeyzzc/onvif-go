package discovery

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type recordingHandler struct {
	entries []slog.Record
}

func (h *recordingHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	h.entries = append(h.entries, r)
	return nil
}
func (h *recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *recordingHandler) WithGroup(string) slog.Handler      { return h }

func (h *recordingHandler) joined() string {
	msgs := make([]string, 0, len(h.entries))
	for _, e := range h.entries {
		msgs = append(msgs, e.Message)
	}
	return strings.Join(msgs, "\n")
}

// DiscoverOptions.Logger emits the probe/discovery lifecycle entries;
// a nil logger (the default) keeps discovery silent (issue #62).
func TestDiscoverWithOptionsLoggerEmitsLifecycleEntries(t *testing.T) {
	rec := &recordingHandler{}

	// Short timeout: on a quiet network no ProbeMatches arrive, which is
	// fine — the probe-sent and summary entries must still be recorded.
	_, _ = DiscoverWithOptions(
		context.Background(),
		300*time.Millisecond,
		&DiscoverOptions{Logger: slog.New(rec)},
	)

	got := rec.joined()
	for _, want := range []string{"ws-discovery probe sent", "ws-discovery complete"} {
		if !strings.Contains(got, want) {
			t.Errorf("log stream missing %q entry; got:\n%s", want, got)
		}
	}
}
