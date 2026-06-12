package lifecycle

import (
	"testing"
	"time"
)

func TestRequestCancellationMarksObservableIntentAndMovesToCancelling(t *testing.T) {
	machine := newRunningLifecycle(t)
	observed := machine.ControlToken()
	occurredAt := time.Date(2026, 6, 12, 11, 0, 0, 0, time.UTC)

	if observed.CancellationRequested() {
		t.Fatal("new running lifecycle should not expose cancellation intent")
	}

	result := machine.RequestCancellation(occurredAt)

	if !result.Accepted {
		t.Fatalf("RequestCancellation rejected: %#v", result.Error)
	}
	if machine.State() != LifecycleStateCancelling {
		t.Fatalf("State = %s, want %s", machine.State(), LifecycleStateCancelling)
	}
	if !observed.CancellationRequested() {
		t.Fatal("previously captured control token should observe cancellation intent")
	}
	assertTransitionRecord(t, result.Record, LifecycleStateRunning, LifecycleStateCancelling, occurredAt, false)
}

func TestCancellationIntentRemainsObservableAfterCancelFinalization(t *testing.T) {
	machine := newRunningLifecycle(t)
	observed := machine.ControlToken()
	cancelAt := time.Date(2026, 6, 12, 11, 5, 0, 0, time.UTC)
	finalizeAt := cancelAt.Add(time.Minute)

	if result := machine.RequestCancellation(cancelAt); !result.Accepted {
		t.Fatalf("RequestCancellation rejected: %#v", result.Error)
	}
	result := machine.TransitionTo(LifecycleStateCancelled, finalizeAt)

	if !result.Accepted {
		t.Fatalf("cancel finalization rejected: %#v", result.Error)
	}
	if machine.State() != LifecycleStateCancelled {
		t.Fatalf("State = %s, want %s", machine.State(), LifecycleStateCancelled)
	}
	if !observed.CancellationRequested() {
		t.Fatal("control token should keep cancellation intent after cancellation reaches terminal state")
	}
}

func TestRequestCancellationRejectsNonRunningLifecycleWithoutMarkingIntent(t *testing.T) {
	machine := NewLifecycle()
	observed := machine.ControlToken()
	occurredAt := time.Date(2026, 6, 12, 11, 10, 0, 0, time.UTC)

	result := machine.RequestCancellation(occurredAt)

	if result.Accepted {
		t.Fatal("INITIALIZED cancellation request should be rejected")
	}
	if machine.State() != LifecycleStateInitialized {
		t.Fatalf("State = %s, want %s", machine.State(), LifecycleStateInitialized)
	}
	if observed.CancellationRequested() {
		t.Fatal("rejected cancellation request must not mark cancellation intent")
	}
	if result.Error == nil || result.Error.Code != LifecycleErrorCodeStateConflict {
		t.Fatalf("expected state conflict error, got %#v", result.Error)
	}
}

func newRunningLifecycle(t *testing.T) *Lifecycle {
	t.Helper()
	machine := NewLifecycle()
	base := time.Date(2026, 6, 12, 10, 55, 0, 0, time.UTC)
	for index, state := range []LifecycleState{
		LifecycleStatePrechecking,
		LifecycleStateReady,
		LifecycleStateRunning,
	} {
		result := machine.TransitionTo(state, base.Add(time.Duration(index)*time.Minute))
		if !result.Accepted {
			t.Fatalf("transition to %s rejected: %#v", state, result.Error)
		}
	}
	return machine
}
