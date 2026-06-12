package lifecycle

import "strings"

// LifecycleErrorCode identifies a stable machine-readable lifecycle error category.
type LifecycleErrorCode string

const (
	// LifecycleErrorCodeRequired reports that a required lifecycle field or prerequisite is missing.
	LifecycleErrorCodeRequired LifecycleErrorCode = "REQUIRED"

	// LifecycleErrorCodeInvalidReference reports that a lifecycle identity or parent reference is invalid.
	LifecycleErrorCodeInvalidReference LifecycleErrorCode = "INVALID_REFERENCE"

	// LifecycleErrorCodeInvalidRange reports that a lifecycle numeric boundary is outside the accepted range.
	LifecycleErrorCodeInvalidRange LifecycleErrorCode = "INVALID_RANGE"

	// LifecycleErrorCodeStateConflict reports that a requested lifecycle state transition is not allowed.
	LifecycleErrorCodeStateConflict LifecycleErrorCode = "STATE_CONFLICT"

	// LifecycleErrorCodeDownstreamFailure reports that a planner, generation, or result seam failed safely.
	LifecycleErrorCodeDownstreamFailure LifecycleErrorCode = "DOWNSTREAM_FAILURE"

	// LifecycleErrorCodeSensitiveValueNotAllowed reports that a candidate public message contained sensitive content.
	LifecycleErrorCodeSensitiveValueNotAllowed LifecycleErrorCode = "SENSITIVE_VALUE_NOT_ALLOWED"
)

// String returns the stable string representation used by lifecycle boundaries.
func (c LifecycleErrorCode) String() string {
	return string(c)
}

// LifecycleStage identifies the lifecycle phase where a safe error summary was produced.
type LifecycleStage string

const (
	// LifecycleStageInputValidation identifies execution input mapping and field-level validation.
	LifecycleStageInputValidation LifecycleStage = "INPUT_VALIDATION"

	// LifecycleStagePrecheck identifies the aggregate precheck phase before execution can start.
	LifecycleStagePrecheck LifecycleStage = "PRECHECK"

	// LifecycleStageStateTransition identifies lifecycle state machine transition checks.
	LifecycleStageStateTransition LifecycleStage = "STATE_TRANSITION"

	// LifecycleStagePlanner identifies the downstream planning seam.
	LifecycleStagePlanner LifecycleStage = "PLANNER"

	// LifecycleStageGeneration identifies the downstream generation seam.
	LifecycleStageGeneration LifecycleStage = "GENERATION"

	// LifecycleStageResult identifies the downstream result-summary seam.
	LifecycleStageResult LifecycleStage = "RESULT"
)

// String returns the stable string representation used by lifecycle boundaries.
func (s LifecycleStage) String() string {
	return string(s)
}

// LifecycleError is the only public lifecycle error result shape and contains a safe summary only.
type LifecycleError struct {
	// Code stores the stable machine-readable lifecycle error category.
	Code LifecycleErrorCode

	// Stage stores the lifecycle phase where the error was produced.
	Stage LifecycleStage

	// FieldPath stores the lifecycle input, state, or seam path related to the error.
	FieldPath string

	// SafeMessage stores a user-readable message without credentials, SQL, connection details, or generated data.
	SafeMessage string
}

// NewLifecycleError creates a lifecycle safe error summary from already-classified error data.
// If the provided message is blank or contains sensitive content, it is replaced with a generic safe message.
func NewLifecycleError(code LifecycleErrorCode, stage LifecycleStage, fieldPath string, safeMessage string) LifecycleError {
	message := strings.TrimSpace(safeMessage)
	if message == "" || containsSensitiveLifecycleContent(message) {
		message = defaultSafeMessage(code)
	}

	return LifecycleError{
		Code:        code,
		Stage:       stage,
		FieldPath:   strings.TrimSpace(fieldPath),
		SafeMessage: message,
	}
}

// MapStateConflictError creates a safe state conflict summary without exposing raw state values.
// The current and requested values classify the conflict only and are never copied into the public message.
func MapStateConflictError(stage LifecycleStage, fieldPath string, current string, requested string) LifecycleError {
	return NewLifecycleError(
		LifecycleErrorCodeStateConflict,
		stage,
		fieldPath,
		"requested lifecycle state transition is not allowed",
	)
}

// MapDownstreamFailure creates a safe downstream failure summary without exposing the raw error payload.
// The raw error is accepted for classification boundaries, but only the code, stage, field path, and generic safe message are returned.
func MapDownstreamFailure(stage LifecycleStage, fieldPath string, err error) LifecycleError {
	return NewLifecycleError(
		LifecycleErrorCodeDownstreamFailure,
		stage,
		fieldPath,
		"downstream lifecycle stage failed",
	)
}

func defaultSafeMessage(code LifecycleErrorCode) string {
	switch code {
	case LifecycleErrorCodeRequired:
		return "required lifecycle field is missing"
	case LifecycleErrorCodeInvalidReference:
		return "lifecycle reference is invalid"
	case LifecycleErrorCodeInvalidRange:
		return "lifecycle numeric boundary is invalid"
	case LifecycleErrorCodeStateConflict:
		return "requested lifecycle state transition is not allowed"
	case LifecycleErrorCodeDownstreamFailure:
		return "downstream lifecycle stage failed"
	case LifecycleErrorCodeSensitiveValueNotAllowed:
		return "lifecycle error message contained sensitive content"
	default:
		return "lifecycle error occurred"
	}
}

func containsSensitiveLifecycleContent(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{
		"password",
		"credential",
		"token",
		"secret",
		"connection string",
		"host=",
		"port=",
		"select ",
		"insert ",
		"update ",
		"delete ",
		"create ",
		"drop ",
		"alter ",
		"user sql",
		"generated data",
		"generateddata",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
