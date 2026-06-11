package execution

import (
	"bytes"
	"encoding/json"
	"errors"
)

// ExecutionTaskStatus identifies the stable task-level execution status.
type ExecutionTaskStatus string

const (
	// ExecutionTaskStatusRunning means the execution task is currently running.
	ExecutionTaskStatusRunning ExecutionTaskStatus = "RUNNING"

	// ExecutionTaskStatusSuccess means every table in the execution task completed successfully.
	ExecutionTaskStatusSuccess ExecutionTaskStatus = "SUCCESS"

	// ExecutionTaskStatusPartialFailed means some tables failed while successfully written data is retained.
	ExecutionTaskStatusPartialFailed ExecutionTaskStatus = "PARTIAL_FAILED"

	// ExecutionTaskStatusFailed means the task failed at task level or all table executions failed.
	ExecutionTaskStatusFailed ExecutionTaskStatus = "FAILED"
)

// IsKnown reports whether the task status belongs to the stable task-level execution status set.
func (s ExecutionTaskStatus) IsKnown() bool {
	switch s {
	case ExecutionTaskStatusRunning,
		ExecutionTaskStatusSuccess,
		ExecutionTaskStatusPartialFailed,
		ExecutionTaskStatusFailed:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the task status is outside the stable task-level execution status set.
func (s ExecutionTaskStatus) IsUnknown() bool {
	return !s.IsKnown()
}

// String returns the stable string representation used for persistence and transport.
// Unknown values are returned unchanged so callers can detect and report them explicitly.
func (s ExecutionTaskStatus) String() string {
	return string(s)
}

// MarshalJSON serializes the task status as its stable string value.
// Unknown values are preserved as strings so validation can classify them explicitly.
func (s ExecutionTaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON restores the task status from its serialized string value.
// Unknown strings are preserved instead of being coerced, allowing validation to reject them explicitly.
func (s *ExecutionTaskStatus) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value == "" && bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("execution task status must be a JSON string")
	}

	*s = ExecutionTaskStatus(value)
	return nil
}

// ExecutionTableStatus identifies the stable table-level execution status.
type ExecutionTableStatus string

const (
	// ExecutionTableStatusPending means the table waits for prerequisite tables to complete.
	ExecutionTableStatusPending ExecutionTableStatus = "PENDING"

	// ExecutionTableStatusRunning means the table execution is currently running.
	ExecutionTableStatusRunning ExecutionTableStatus = "RUNNING"

	// ExecutionTableStatusSuccess means rows were written successfully for the table.
	ExecutionTableStatusSuccess ExecutionTableStatus = "SUCCESS"

	// ExecutionTableStatusFailed means writing rows for the table failed.
	ExecutionTableStatusFailed ExecutionTableStatus = "FAILED"

	// ExecutionTableStatusSkipped means the table was skipped because a prerequisite table failed.
	ExecutionTableStatusSkipped ExecutionTableStatus = "SKIPPED"
)

// IsKnown reports whether the table status belongs to the stable table-level execution status set.
func (s ExecutionTableStatus) IsKnown() bool {
	switch s {
	case ExecutionTableStatusPending,
		ExecutionTableStatusRunning,
		ExecutionTableStatusSuccess,
		ExecutionTableStatusFailed,
		ExecutionTableStatusSkipped:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the table status is outside the stable table-level execution status set.
func (s ExecutionTableStatus) IsUnknown() bool {
	return !s.IsKnown()
}

// String returns the stable string representation used for persistence and transport.
// Unknown values are returned unchanged so callers can detect and report them explicitly.
func (s ExecutionTableStatus) String() string {
	return string(s)
}

// MarshalJSON serializes the table status as its stable string value.
// Unknown values are preserved as strings so validation can classify them explicitly.
func (s ExecutionTableStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON restores the table status from its serialized string value.
// Unknown strings are preserved instead of being coerced, allowing validation to reject them explicitly.
func (s *ExecutionTableStatus) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value == "" && bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("execution table status must be a JSON string")
	}

	*s = ExecutionTableStatus(value)
	return nil
}
