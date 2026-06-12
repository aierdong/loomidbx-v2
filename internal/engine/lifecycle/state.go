package lifecycle

import "time"

// LifecycleState identifies the engine-internal runtime state for one lifecycle instance.
type LifecycleState string

const (
	// LifecycleStateInitialized means the lifecycle can be prechecked but has not started running.
	LifecycleStateInitialized LifecycleState = "INITIALIZED"

	// LifecycleStatePrechecking means startup prerequisites are being evaluated and the lifecycle is not running.
	LifecycleStatePrechecking LifecycleState = "PRECHECKING"

	// LifecycleStateReady means precheck passed and the lifecycle may be started by the coordinator.
	LifecycleStateReady LifecycleState = "READY"

	// LifecycleStateRunning means execution has started and downstream engine steps may observe runtime state.
	LifecycleStateRunning LifecycleState = "RUNNING"

	// LifecycleStateCancelling means cancellation has been requested and final cancellation has not yet been recorded.
	LifecycleStateCancelling LifecycleState = "CANCELLING"

	// LifecycleStateCancelled means execution ended through the cancellation path and cannot be controlled again.
	LifecycleStateCancelled LifecycleState = "CANCELLED"

	// LifecycleStateFailed means execution ended through the failure path and cannot be controlled again.
	LifecycleStateFailed LifecycleState = "FAILED"

	// LifecycleStateCompleted means execution ended successfully and cannot be controlled again.
	LifecycleStateCompleted LifecycleState = "COMPLETED"
)

// String returns the stable string representation used by lifecycle boundaries.
func (s LifecycleState) String() string {
	return string(s)
}

// IsTerminal reports whether the state is one of the mutually exclusive lifecycle terminal states.
func (s LifecycleState) IsTerminal() bool {
	switch s {
	case LifecycleStateCancelled, LifecycleStateFailed, LifecycleStateCompleted:
		return true
	default:
		return false
	}
}

// TransitionRecord stores one accepted or rejected lifecycle state transition attempt.
type TransitionRecord struct {
	// From stores the lifecycle state observed before the transition attempt.
	From LifecycleState

	// To stores the lifecycle state requested by the transition attempt.
	To LifecycleState

	// Stage stores the lifecycle stage that evaluated the transition.
	Stage LifecycleStage

	// OccurredAt stores when the transition attempt was made.
	OccurredAt time.Time

	// Rejected reports whether the transition was refused by state machine rules.
	Rejected bool

	// Rejection stores a safe conflict summary when Rejected is true.
	Rejection *LifecycleError
}

// TransitionResult summarizes the outcome of a lifecycle state transition attempt.
type TransitionResult struct {
	// Accepted reports whether the lifecycle state changed to the requested target state.
	Accepted bool

	// Record stores the transition attempt record appended to the lifecycle history.
	Record TransitionRecord

	// Error stores the safe state conflict summary when Accepted is false.
	Error *LifecycleError
}

// Lifecycle stores the engine-internal state machine for a single execution lifecycle.
type Lifecycle struct {
	// state stores the current engine-internal lifecycle state.
	state LifecycleState

	// records stores accepted and rejected state transition attempts in chronological order.
	records []TransitionRecord
}

// NewLifecycle creates a lifecycle in the initialized state, which is precheckable but not running.
func NewLifecycle() *Lifecycle {
	return &Lifecycle{state: LifecycleStateInitialized}
}

// State returns the current engine-internal lifecycle state.
func (l *Lifecycle) State() LifecycleState {
	if l == nil {
		return ""
	}
	return l.state
}

// CanPrecheck reports whether the lifecycle is in the initial state that allows precheck to begin.
func (l *Lifecycle) CanPrecheck() bool {
	return l != nil && l.state == LifecycleStateInitialized
}

// TransitionRecords returns a defensive copy of accepted and rejected lifecycle transition records.
func (l *Lifecycle) TransitionRecords() []TransitionRecord {
	if l == nil || len(l.records) == 0 {
		return nil
	}
	records := make([]TransitionRecord, len(l.records))
	copy(records, l.records)
	return records
}

// TransitionTo attempts to move the lifecycle to a requested state and records accepted or rejected results.
// It returns a state conflict error when the transition is not allowed or the lifecycle is already terminal.
func (l *Lifecycle) TransitionTo(to LifecycleState, occurredAt time.Time) TransitionResult {
	if l == nil {
		conflict := MapStateConflictError(LifecycleStageStateTransition, "state", "", to.String())
		record := TransitionRecord{
			To:         to,
			Stage:      LifecycleStageStateTransition,
			OccurredAt: occurredAt,
			Rejected:   true,
			Rejection:  &conflict,
		}
		return TransitionResult{Accepted: false, Record: record, Error: &conflict}
	}

	from := l.state
	record := TransitionRecord{
		From:       from,
		To:         to,
		Stage:      LifecycleStageStateTransition,
		OccurredAt: occurredAt,
	}

	if !isAllowedLifecycleTransition(from, to) {
		conflict := MapStateConflictError(LifecycleStageStateTransition, "state", from.String(), to.String())
		record.Rejected = true
		record.Rejection = &conflict
		l.records = append(l.records, record)
		return TransitionResult{Accepted: false, Record: record, Error: &conflict}
	}

	l.state = to
	l.records = append(l.records, record)
	return TransitionResult{Accepted: true, Record: record}
}

func isAllowedLifecycleTransition(from LifecycleState, to LifecycleState) bool {
	if from.IsTerminal() {
		return false
	}

	switch from {
	case LifecycleStateInitialized:
		return to == LifecycleStatePrechecking
	case LifecycleStatePrechecking:
		return to == LifecycleStateReady || to == LifecycleStateFailed
	case LifecycleStateReady:
		return to == LifecycleStateRunning
	case LifecycleStateRunning:
		return to == LifecycleStateCancelling || to == LifecycleStateFailed || to == LifecycleStateCompleted
	case LifecycleStateCancelling:
		return to == LifecycleStateCancelled || to == LifecycleStateFailed
	default:
		return false
	}
}
