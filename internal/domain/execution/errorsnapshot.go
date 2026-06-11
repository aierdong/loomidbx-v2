package execution

import "time"

// ExecutionErrorSnapshot stores a safe, displayable error summary for execution history.
type ExecutionErrorSnapshot struct {
	// Code stores a stable machine-readable error code.
	Code string `json:"code"`

	// Message stores a safe user-readable error summary without credentials, user SQL, or generated data.
	Message string `json:"message"`

	// FieldPath stores an optional lower camelCase field path related to the error.
	FieldPath string `json:"fieldPath,omitempty"`

	// OccurredAt stores when the execution error occurred.
	OccurredAt time.Time `json:"occurredAt"`
}
