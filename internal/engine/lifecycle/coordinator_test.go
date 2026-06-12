package lifecycle

import (
	"testing"
	"time"
)

func TestCoordinatorRunsValidSnapshotToCompletedLifecycle(t *testing.T) {
	job := validGenerationJobSnapshot()
	precheckedAt := time.Date(2026, 6, 12, 11, 58, 0, 0, time.UTC)
	readyAt := time.Date(2026, 6, 12, 11, 59, 0, 0, time.UTC)
	startedAt := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
	completedAt := startedAt.Add(2 * time.Minute)
	coordinator := NewLifecycleCoordinator(sequenceClock(precheckedAt, readyAt, startedAt, completedAt))

	result := coordinator.Run(job)

	if result.Input == nil {
		t.Fatal("Run should expose mapped execution input")
	}
	if !result.Precheck.Passed {
		t.Fatalf("Precheck.Passed = false, blocking=%#v", result.Precheck.BlockingErrors)
	}
	if result.State != LifecycleStateCompleted {
		t.Fatalf("State = %s, want %s", result.State, LifecycleStateCompleted)
	}
	if result.StartedAt == nil || !result.StartedAt.Equal(startedAt) {
		t.Fatalf("StartedAt = %v, want %s", result.StartedAt, startedAt)
	}
	if result.CompletedAt == nil || !result.CompletedAt.Equal(completedAt) {
		t.Fatalf("CompletedAt = %v, want %s", result.CompletedAt, completedAt)
	}
	if result.Control.CancellationRequested() {
		t.Fatal("completed lifecycle should not mark cancellation intent")
	}
	if result.Failure != nil {
		t.Fatalf("Failure = %#v, want nil", result.Failure)
	}
	assertTransitionPath(t, result.Transitions, []LifecycleState{
		LifecycleStateInitialized,
		LifecycleStatePrechecking,
		LifecycleStateReady,
		LifecycleStateRunning,
		LifecycleStateCompleted,
	})
}

func TestCoordinatorStopsInvalidSnapshotBeforeRunning(t *testing.T) {
	job := validGenerationJobSnapshot()
	job.Task.ID = 0
	startedAt := time.Date(2026, 6, 12, 12, 10, 0, 0, time.UTC)
	coordinator := NewLifecycleCoordinator(sequenceClock(startedAt))

	result := coordinator.Run(job)

	if result.Input != nil {
		t.Fatalf("Input = %#v, want nil", result.Input)
	}
	if result.Precheck.Passed {
		t.Fatal("invalid snapshot should fail precheck")
	}
	if result.State != LifecycleStateFailed {
		t.Fatalf("State = %s, want %s", result.State, LifecycleStateFailed)
	}
	if result.StartedAt != nil {
		t.Fatalf("StartedAt = %v, want nil", result.StartedAt)
	}
	if result.CompletedAt != nil {
		t.Fatalf("CompletedAt = %v, want nil", result.CompletedAt)
	}
	if result.Failure == nil {
		t.Fatal("precheck failure should expose safe failure summary")
	}
	if result.Failure.Stage != LifecycleStagePrecheck {
		t.Fatalf("Failure.Stage = %s, want %s", result.Failure.Stage, LifecycleStagePrecheck)
	}
	assertTransitionPath(t, result.Transitions, []LifecycleState{
		LifecycleStateInitialized,
		LifecycleStatePrechecking,
		LifecycleStateFailed,
	})
}

func TestCoordinatorResultTransitionRecordsAreDefensiveCopies(t *testing.T) {
	job := validGenerationJobSnapshot()
	coordinator := NewLifecycleCoordinator(sequenceClock(
		time.Date(2026, 6, 12, 12, 20, 0, 0, time.UTC),
		time.Date(2026, 6, 12, 12, 21, 0, 0, time.UTC),
	))

	result := coordinator.Run(job)
	result.Transitions[0].To = LifecycleStateFailed

	again := coordinator.Run(job)
	if again.Transitions[0].To != LifecycleStatePrechecking {
		t.Fatal("coordinator results should not share transition record backing storage")
	}
}

func sequenceClock(times ...time.Time) func() time.Time {
	index := 0
	return func() time.Time {
		if index >= len(times) {
			return times[len(times)-1]
		}
		value := times[index]
		index++
		return value
	}
}

func assertTransitionPath(t *testing.T, records []TransitionRecord, states []LifecycleState) {
	t.Helper()
	if len(records) != len(states)-1 {
		t.Fatalf("transition count = %d, want %d", len(records), len(states)-1)
	}
	for index, record := range records {
		if record.From != states[index] || record.To != states[index+1] || record.Rejected {
			t.Fatalf("record[%d] = %s -> %s rejected=%t, want %s -> %s accepted", index, record.From, record.To, record.Rejected, states[index], states[index+1])
		}
	}
}
