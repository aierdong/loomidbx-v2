package repository

import "fmt"

// ErrorCode identifies stable repository error categories.
type ErrorCode string

const (
	// ErrorCodeNotImplemented reports a repository capability reserved for a later business spec.
	ErrorCodeNotImplemented ErrorCode = "REPOSITORY_NOT_IMPLEMENTED"

	// ErrorCodeNotMigrated reports a repository capability whose required migration has not been applied.
	ErrorCodeNotMigrated ErrorCode = "REPOSITORY_NOT_MIGRATED"

	// ErrorCodeStorageUnavailable reports repository access before storage is initialized.
	ErrorCodeStorageUnavailable ErrorCode = "REPOSITORY_STORAGE_UNAVAILABLE"
)

// RepositoryError carries sanitized repository failure context.
type RepositoryError struct {
	// Code is the stable machine-readable error category.
	Code ErrorCode

	// Capability is the repository capability or boundary related to the failure.
	Capability string

	// Operation is the repository operation that failed.
	Operation string

	// Reason is a sanitized human-readable reason.
	Reason string

	// Err is the wrapped cause for Go error chaining.
	Err error
}

// Error returns a sanitized repository error message.
func (err *RepositoryError) Error() string {
	if err == nil {
		return "repository error"
	}
	message := "repository " + string(err.Code)
	if err.Operation != "" {
		message += " during " + err.Operation
	}
	if err.Capability != "" {
		message += " for " + err.Capability
	}
	if err.Reason != "" {
		message += ": " + err.Reason
	}
	return message
}

// Unwrap returns the wrapped cause.
func (err *RepositoryError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

// Is matches repository errors by stable category.
func (err *RepositoryError) Is(target error) bool {
	targetErr, ok := target.(*RepositoryError)
	if !ok || err == nil || targetErr == nil {
		return false
	}
	return targetErr.Code == "" || err.Code == targetErr.Code
}

// NewError constructs a repository error with non-sensitive context.
func NewError(code ErrorCode, capability string, operation string, reason string, cause error) *RepositoryError {
	var wrapped error
	if cause != nil {
		wrapped = fmt.Errorf("repository cause redacted")
	}
	return &RepositoryError{Code: code, Capability: capability, Operation: operation, Reason: reason, Err: wrapped}
}
