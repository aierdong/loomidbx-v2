package lifecycle

import (
	"strings"
	"time"

	domainexecution "github.com/gerdong/loomidbx/internal/domain/execution"
)

// LifecycleSnapshot is the lifecycle-owned current or final state view for later history, API, or UI boundaries.
type LifecycleSnapshot struct {
	// State stores the current or final engine-internal lifecycle state.
	State LifecycleState

	// StartedAt stores when execution entered RUNNING; nil means the lifecycle never started.
	StartedAt *time.Time

	// EndedAt stores when a terminal lifecycle state was recorded; nil means the lifecycle has not ended.
	EndedAt *time.Time

	// CancellationRequested reports whether cancellation intent was recorded before the snapshot was created.
	CancellationRequested bool

	// Failure stores a safe failure summary for failed final snapshots.
	Failure *LifecycleError

	// Transitions stores a defensive copy of accepted and rejected lifecycle transition records.
	Transitions []TransitionRecord
}

// NewLifecycleSnapshot creates a lifecycle snapshot with defensive copies of mutable times, failure summary, and transition records.
func NewLifecycleSnapshot(state LifecycleState, startedAt *time.Time, endedAt *time.Time, cancellationRequested bool, failure *LifecycleError, transitions []TransitionRecord) LifecycleSnapshot {
	return LifecycleSnapshot{
		State:                 state,
		StartedAt:             cloneTimePointer(startedAt),
		EndedAt:               cloneTimePointer(endedAt),
		CancellationRequested: cancellationRequested,
		Failure:               cloneLifecycleErrorPointer(failure),
		Transitions:           cloneTransitionRecords(transitions),
	}
}

// TransitionRecords returns a defensive copy of the lifecycle transition records in this snapshot.
func (s LifecycleSnapshot) TransitionRecords() []TransitionRecord {
	return cloneTransitionRecords(s.Transitions)
}

// Phase2HistoryStatus maps only final lifecycle snapshots to Phase 2 execution history statuses.
// It deliberately does not add lifecycle runtime states to the Phase 2 status enum.
func (s LifecycleSnapshot) Phase2HistoryStatus() (domainexecution.ExecutionTaskStatus, bool) {
	switch s.State {
	case LifecycleStateCompleted:
		return domainexecution.ExecutionTaskStatusSuccess, true
	case LifecycleStateFailed:
		return domainexecution.ExecutionTaskStatusFailed, true
	case LifecycleStateCancelled:
		return domainexecution.ExecutionTaskStatusFailed, true
	default:
		return "", false
	}
}

func cloneTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func cloneLifecycleErrorPointer(value *LifecycleError) *LifecycleError {
	if value == nil {
		return nil
	}
	code := value.Code
	if strings.TrimSpace(code.String()) == "" || containsSensitiveLifecycleContent(code.String()) {
		code = LifecycleErrorCodeSensitiveValueNotAllowed
	}
	stage := value.Stage
	if containsSensitiveLifecycleContent(stage.String()) {
		stage = LifecycleStageStateTransition
	}
	fieldPath := value.FieldPath
	if containsSensitiveLifecycleContent(fieldPath) {
		fieldPath = ""
	}
	clone := NewLifecycleError(code, stage, fieldPath, value.SafeMessage)
	return &clone
}

func cloneTransitionRecords(records []TransitionRecord) []TransitionRecord {
	if len(records) == 0 {
		return nil
	}
	clone := make([]TransitionRecord, len(records))
	copy(clone, records)
	for index := range clone {
		clone[index].Rejection = cloneLifecycleErrorPointer(clone[index].Rejection)
	}
	return clone
}
