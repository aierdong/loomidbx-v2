package lifecycle

import (
	"strings"
	"testing"
	"time"
)

func TestNewLifecycleStartsInitializedAndPrecheckable(t *testing.T) {
	machine := NewLifecycle()

	if machine.State() != LifecycleStateInitialized {
		t.Fatalf("State = %s, want %s", machine.State(), LifecycleStateInitialized)
	}
	if !machine.CanPrecheck() {
		t.Fatal("new lifecycle should be precheckable")
	}
	if machine.State().IsTerminal() {
		t.Fatal("initialized lifecycle must not be terminal")
	}
	if records := machine.TransitionRecords(); len(records) != 0 {
		t.Fatalf("TransitionRecords length = %d, want 0", len(records))
	}
}

func TestLifecycleStateTerminalDetection(t *testing.T) {
	terminalStates := []LifecycleState{
		LifecycleStateCancelled,
		LifecycleStateFailed,
		LifecycleStateCompleted,
	}
	for _, state := range terminalStates {
		if !state.IsTerminal() {
			t.Fatalf("%s should be terminal", state)
		}
	}

	nonTerminalStates := []LifecycleState{
		LifecycleStateInitialized,
		LifecycleStatePrechecking,
		LifecycleStateReady,
		LifecycleStateRunning,
		LifecycleStateCancelling,
	}
	for _, state := range nonTerminalStates {
		if state.IsTerminal() {
			t.Fatalf("%s should not be terminal", state)
		}
	}
}

func TestLifecycleTransitionRecordsAcceptedFlow(t *testing.T) {
	machine := NewLifecycle()
	occurredAt := time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC)

	result := machine.TransitionTo(LifecycleStatePrechecking, occurredAt)

	if !result.Accepted {
		t.Fatalf("TransitionTo PRECHECKING rejected: %#v", result.Error)
	}
	if result.Error != nil {
		t.Fatalf("accepted transition returned error: %#v", result.Error)
	}
	if machine.State() != LifecycleStatePrechecking {
		t.Fatalf("State = %s, want %s", machine.State(), LifecycleStatePrechecking)
	}
	assertTransitionRecord(t, result.Record, LifecycleStateInitialized, LifecycleStatePrechecking, occurredAt, false)

	records := machine.TransitionRecords()
	if len(records) != 1 {
		t.Fatalf("TransitionRecords length = %d, want 1", len(records))
	}
	assertTransitionRecord(t, records[0], LifecycleStateInitialized, LifecycleStatePrechecking, occurredAt, false)

	records[0].To = LifecycleStateCompleted
	if machine.TransitionRecords()[0].To != LifecycleStatePrechecking {
		t.Fatal("TransitionRecords should return a defensive copy")
	}
}

func TestLifecycleTransitionRejectsIllegalFlowWithoutChangingState(t *testing.T) {
	machine := NewLifecycle()
	occurredAt := time.Date(2026, 6, 12, 10, 5, 0, 0, time.UTC)

	result := machine.TransitionTo(LifecycleStateRunning, occurredAt)

	if result.Accepted {
		t.Fatal("INITIALIZED -> RUNNING should be rejected")
	}
	if machine.State() != LifecycleStateInitialized {
		t.Fatalf("State changed after rejected transition: %s", machine.State())
	}
	if result.Error == nil {
		t.Fatal("rejected transition should return a state conflict error")
	}
	if result.Error.Code != LifecycleErrorCodeStateConflict {
		t.Fatalf("Error.Code = %s, want %s", result.Error.Code, LifecycleErrorCodeStateConflict)
	}
	if result.Error.Stage != LifecycleStageStateTransition {
		t.Fatalf("Error.Stage = %s, want %s", result.Error.Stage, LifecycleStageStateTransition)
	}
	if result.Error.FieldPath != "state" {
		t.Fatalf("Error.FieldPath = %q, want state", result.Error.FieldPath)
	}
	if strings.Contains(result.Error.SafeMessage, string(LifecycleStateInitialized)) || strings.Contains(result.Error.SafeMessage, string(LifecycleStateRunning)) {
		t.Fatalf("state conflict message leaked raw states: %q", result.Error.SafeMessage)
	}
	assertTransitionRecord(t, result.Record, LifecycleStateInitialized, LifecycleStateRunning, occurredAt, true)
	if result.Record.Rejection == nil {
		t.Fatal("rejected record should keep the safe rejection summary")
	}

	records := machine.TransitionRecords()
	if len(records) != 1 {
		t.Fatalf("TransitionRecords length = %d, want 1 rejected record", len(records))
	}
	assertTransitionRecord(t, records[0], LifecycleStateInitialized, LifecycleStateRunning, occurredAt, true)
}

func TestLifecycleTerminalStateRejectsFurtherTransitions(t *testing.T) {
	machine := NewLifecycle()
	base := time.Date(2026, 6, 12, 10, 10, 0, 0, time.UTC)
	for index, state := range []LifecycleState{
		LifecycleStatePrechecking,
		LifecycleStateReady,
		LifecycleStateRunning,
		LifecycleStateCompleted,
	} {
		result := machine.TransitionTo(state, base.Add(time.Duration(index)*time.Minute))
		if !result.Accepted {
			t.Fatalf("transition to %s rejected: %#v", state, result.Error)
		}
	}
	if !machine.State().IsTerminal() {
		t.Fatalf("State = %s, want terminal", machine.State())
	}

	result := machine.TransitionTo(LifecycleStateFailed, base.Add(10*time.Minute))

	if result.Accepted {
		t.Fatal("terminal COMPLETED -> FAILED should be rejected")
	}
	if machine.State() != LifecycleStateCompleted {
		t.Fatalf("State = %s, want %s", machine.State(), LifecycleStateCompleted)
	}
	if result.Error == nil || result.Error.Code != LifecycleErrorCodeStateConflict {
		t.Fatalf("expected terminal transition conflict error, got %#v", result.Error)
	}
}

func assertTransitionRecord(t *testing.T, record TransitionRecord, from LifecycleState, to LifecycleState, occurredAt time.Time, rejected bool) {
	t.Helper()
	if record.From != from {
		t.Fatalf("record.From = %s, want %s", record.From, from)
	}
	if record.To != to {
		t.Fatalf("record.To = %s, want %s", record.To, to)
	}
	if !record.OccurredAt.Equal(occurredAt) {
		t.Fatalf("record.OccurredAt = %s, want %s", record.OccurredAt, occurredAt)
	}
	if record.Stage != LifecycleStageStateTransition {
		t.Fatalf("record.Stage = %s, want %s", record.Stage, LifecycleStageStateTransition)
	}
	if record.Rejected != rejected {
		t.Fatalf("record.Rejected = %t, want %t", record.Rejected, rejected)
	}
}
