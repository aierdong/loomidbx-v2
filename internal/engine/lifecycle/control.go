package lifecycle

import (
	"sync/atomic"
	"time"
)

// ControlToken exposes runtime control intent to downstream lifecycle steps.
type ControlToken struct {
	requested *atomic.Bool
}

// CancellationRequested reports whether cancellation has been requested for the lifecycle.
func (t ControlToken) CancellationRequested() bool {
	return t.requested != nil && t.requested.Load()
}

// ControlToken returns a stable observable control token for downstream lifecycle steps.
func (l *Lifecycle) ControlToken() ControlToken {
	if l == nil {
		return ControlToken{}
	}
	return ControlToken{requested: &l.cancellationRequested}
}

// RequestCancellation marks cancellation intent and moves a running lifecycle onto the cancellation path.
func (l *Lifecycle) RequestCancellation(occurredAt time.Time) TransitionResult {
	result := l.TransitionTo(LifecycleStateCancelling, occurredAt)
	if result.Accepted {
		l.cancellationRequested.Store(true)
	}
	return result
}
