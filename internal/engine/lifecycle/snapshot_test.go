package lifecycle

import (
	"reflect"
	"testing"
	"time"

	domainexecution "github.com/gerdong/loomidbx/internal/domain/execution"
)

func TestLifecycleSnapshotCapturesFinalBoundaryState(t *testing.T) {
	startedAt := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	endedAt := startedAt.Add(3 * time.Minute)
	failure := NewLifecycleError(LifecycleErrorCodeDownstreamFailure, LifecycleStageGeneration, "generation.batch", "password=secret SELECT * FROM users")
	transitionAt := startedAt.Add(time.Minute)
	transitions := []TransitionRecord{
		{
			From:       LifecycleStateRunning,
			To:         LifecycleStateFailed,
			Stage:      LifecycleStageStateTransition,
			OccurredAt: transitionAt,
		},
	}

	snapshot := NewLifecycleSnapshot(LifecycleStateFailed, &startedAt, &endedAt, true, &failure, transitions)

	if snapshot.State != LifecycleStateFailed {
		t.Fatalf("State = %s, want %s", snapshot.State, LifecycleStateFailed)
	}
	if snapshot.StartedAt == nil || !snapshot.StartedAt.Equal(startedAt) {
		t.Fatalf("StartedAt = %v, want %s", snapshot.StartedAt, startedAt)
	}
	if snapshot.EndedAt == nil || !snapshot.EndedAt.Equal(endedAt) {
		t.Fatalf("EndedAt = %v, want %s", snapshot.EndedAt, endedAt)
	}
	if !snapshot.CancellationRequested {
		t.Fatal("CancellationRequested = false, want true")
	}
	assertDownstreamFailureSummary(t, snapshot.Failure, LifecycleStageGeneration, "generation.batch")
	assertTransitionRecord(t, snapshot.Transitions[0], LifecycleStateRunning, LifecycleStateFailed, transitionAt, false)
}

func TestLifecycleSnapshotReturnsDefensiveCopies(t *testing.T) {
	startedAt := time.Date(2026, 6, 13, 9, 10, 0, 0, time.UTC)
	endedAt := startedAt.Add(time.Minute)
	failure := NewLifecycleError(LifecycleErrorCodeDownstreamFailure, LifecycleStageResult, "result.summary", "summary failed safely")
	transitions := []TransitionRecord{
		{From: LifecycleStateRunning, To: LifecycleStateFailed, Stage: LifecycleStageStateTransition, OccurredAt: endedAt},
	}

	snapshot := NewLifecycleSnapshot(LifecycleStateFailed, &startedAt, &endedAt, false, &failure, transitions)
	startedAt = startedAt.Add(24 * time.Hour)
	endedAt = endedAt.Add(24 * time.Hour)
	failure.FieldPath = "mutated"
	transitions[0].To = LifecycleStateCompleted

	if snapshot.StartedAt.Equal(startedAt) {
		t.Fatal("StartedAt should be copied from caller-owned time")
	}
	if snapshot.EndedAt.Equal(endedAt) {
		t.Fatal("EndedAt should be copied from caller-owned time")
	}
	if snapshot.Failure.FieldPath != "result.summary" {
		t.Fatalf("Failure.FieldPath = %q, want defensive copy", snapshot.Failure.FieldPath)
	}
	if snapshot.Transitions[0].To != LifecycleStateFailed {
		t.Fatalf("Transitions[0].To = %s, want defensive copy", snapshot.Transitions[0].To)
	}

	records := snapshot.TransitionRecords()
	records[0].To = LifecycleStateCompleted
	if snapshot.TransitionRecords()[0].To != LifecycleStateFailed {
		t.Fatal("TransitionRecords should return defensive copies")
	}
}

func TestLifecycleSnapshotMapsOnlyFinalStatesToPhase2HistoryStatus(t *testing.T) {
	tests := []struct {
		state      LifecycleState
		wantStatus domainexecution.ExecutionTaskStatus
		wantOK     bool
	}{
		{state: LifecycleStateCompleted, wantStatus: domainexecution.ExecutionTaskStatusSuccess, wantOK: true},
		{state: LifecycleStateFailed, wantStatus: domainexecution.ExecutionTaskStatusFailed, wantOK: true},
		{state: LifecycleStateCancelled, wantStatus: domainexecution.ExecutionTaskStatusFailed, wantOK: true},
		{state: LifecycleStateInitialized, wantOK: false},
		{state: LifecycleStatePrechecking, wantOK: false},
		{state: LifecycleStateReady, wantOK: false},
		{state: LifecycleStateRunning, wantOK: false},
		{state: LifecycleStateCancelling, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			snapshot := NewLifecycleSnapshot(tt.state, nil, nil, false, nil, nil)

			status, ok := snapshot.Phase2HistoryStatus()

			if ok != tt.wantOK {
				t.Fatalf("ok = %t, want %t", ok, tt.wantOK)
			}
			if status != tt.wantStatus {
				t.Fatalf("status = %s, want %s", status, tt.wantStatus)
			}
		})
	}
}

func TestLifecycleSnapshotRepresentsCancellationTerminalState(t *testing.T) {
	startedAt := time.Date(2026, 6, 13, 9, 25, 0, 0, time.UTC)
	endedAt := startedAt.Add(90 * time.Second)
	transitions := []TransitionRecord{
		{From: LifecycleStateInitialized, To: LifecycleStatePrechecking, Stage: LifecycleStageStateTransition, OccurredAt: startedAt.Add(-2 * time.Minute)},
		{From: LifecycleStatePrechecking, To: LifecycleStateReady, Stage: LifecycleStageStateTransition, OccurredAt: startedAt.Add(-time.Minute)},
		{From: LifecycleStateReady, To: LifecycleStateRunning, Stage: LifecycleStageStateTransition, OccurredAt: startedAt},
		{From: LifecycleStateRunning, To: LifecycleStateCancelling, Stage: LifecycleStageStateTransition, OccurredAt: startedAt.Add(time.Minute)},
		{From: LifecycleStateCancelling, To: LifecycleStateCancelled, Stage: LifecycleStageStateTransition, OccurredAt: endedAt},
	}

	snapshot := NewLifecycleSnapshot(LifecycleStateCancelled, &startedAt, &endedAt, true, nil, transitions)

	if snapshot.State != LifecycleStateCancelled {
		t.Fatalf("State = %s, want %s", snapshot.State, LifecycleStateCancelled)
	}
	if snapshot.StartedAt == nil || !snapshot.StartedAt.Equal(startedAt) {
		t.Fatalf("StartedAt = %v, want %s", snapshot.StartedAt, startedAt)
	}
	if snapshot.EndedAt == nil || !snapshot.EndedAt.Equal(endedAt) {
		t.Fatalf("EndedAt = %v, want %s", snapshot.EndedAt, endedAt)
	}
	if !snapshot.CancellationRequested {
		t.Fatal("CancellationRequested = false, want true")
	}
	if snapshot.Failure != nil {
		t.Fatalf("Failure = %#v, want nil", snapshot.Failure)
	}
	assertTransitionPath(t, snapshot.Transitions, []LifecycleState{
		LifecycleStateInitialized,
		LifecycleStatePrechecking,
		LifecycleStateReady,
		LifecycleStateRunning,
		LifecycleStateCancelling,
		LifecycleStateCancelled,
	})
}

func TestCoordinatorResultProvidesFinalLifecycleSnapshot(t *testing.T) {
	startedAt := time.Date(2026, 6, 13, 9, 20, 0, 0, time.UTC)
	completedAt := startedAt.Add(2 * time.Minute)
	coordinator := NewLifecycleCoordinator(sequenceClock(
		startedAt.Add(-2*time.Minute),
		startedAt.Add(-time.Minute),
		startedAt,
		completedAt,
	))

	result := coordinator.Run(validGenerationJobSnapshot())
	snapshot := result.Snapshot

	if snapshot.State != LifecycleStateCompleted {
		t.Fatalf("snapshot.State = %s, want %s", snapshot.State, LifecycleStateCompleted)
	}
	if snapshot.StartedAt == nil || !snapshot.StartedAt.Equal(startedAt) {
		t.Fatalf("snapshot.StartedAt = %v, want %s", snapshot.StartedAt, startedAt)
	}
	if snapshot.EndedAt == nil || !snapshot.EndedAt.Equal(completedAt) {
		t.Fatalf("snapshot.EndedAt = %v, want %s", snapshot.EndedAt, completedAt)
	}
	if snapshot.CancellationRequested {
		t.Fatal("snapshot.CancellationRequested = true, want false")
	}
	if snapshot.Failure != nil {
		t.Fatalf("snapshot.Failure = %#v, want nil", snapshot.Failure)
	}
	if !reflect.DeepEqual(snapshot.Transitions, result.Transitions) {
		t.Fatalf("snapshot.Transitions = %#v, want result transitions %#v", snapshot.Transitions, result.Transitions)
	}
	status, ok := snapshot.Phase2HistoryStatus()
	if !ok || status != domainexecution.ExecutionTaskStatusSuccess {
		t.Fatalf("Phase2HistoryStatus = %s, %t; want %s, true", status, ok, domainexecution.ExecutionTaskStatusSuccess)
	}
}
