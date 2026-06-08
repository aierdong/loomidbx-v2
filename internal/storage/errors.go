package storage

import "fmt"

// ErrorCode identifies stable storage error categories for service and facade conversion.
type ErrorCode string

const (
	// ErrorCodePathInvalid reports an unusable data directory or layout path.
	ErrorCodePathInvalid ErrorCode = "STORAGE_PATH_INVALID"

	// ErrorCodeOpenFailed reports failure to open the local structured data file.
	ErrorCodeOpenFailed ErrorCode = "STORAGE_OPEN_FAILED"

	// ErrorCodeMigrationFailed reports failure while validating or applying migrations.
	ErrorCodeMigrationFailed ErrorCode = "STORAGE_MIGRATION_FAILED"

	// ErrorCodeSecretStoreUnavailable reports that no secure secret store implementation is configured.
	ErrorCodeSecretStoreUnavailable ErrorCode = "SECRET_STORE_UNAVAILABLE"
)

// StorageError carries sanitized storage failure context.
type StorageError struct {
	// Code is the stable machine-readable error category.
	Code ErrorCode

	// Field is the config, storage, or secret field related to the failure.
	Field string

	// Operation is the storage operation that failed.
	Operation string

	// MigrationVersion is the migration version related to a migration failure.
	MigrationVersion string

	// Reason is a sanitized human-readable reason and must not contain sensitive values.
	Reason string

	// Err is the wrapped cause for Go error chaining.
	Err error
}

// Error returns a sanitized storage error message.
func (err *StorageError) Error() string {
	if err == nil {
		return "storage error"
	}
	message := "storage " + string(err.Code)
	if err.Operation != "" {
		message += " during " + err.Operation
	}
	if err.Field != "" {
		message += " on " + err.Field
	}
	if err.MigrationVersion != "" {
		message += " for migration " + err.MigrationVersion
	}
	if err.Reason != "" {
		message += ": " + err.Reason
	}
	return message
}

// Unwrap returns the wrapped non-sensitive cause.
func (err *StorageError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

// Is matches storage errors by stable category.
func (err *StorageError) Is(target error) bool {
	targetErr, ok := target.(*StorageError)
	if !ok || err == nil || targetErr == nil {
		return false
	}
	return targetErr.Code == "" || err.Code == targetErr.Code
}

// NewError constructs a typed storage error with sanitized context.
func NewError(code ErrorCode, field string, operation string, reason string, cause error) *StorageError {
	return NewStorageError(string(code), field, operation, "", reason, cause)
}

// NewStorageError constructs a typed storage error from primitive strings for leaf packages.
func NewStorageError(code string, field string, operation string, migrationVersion string, reason string, cause error) *StorageError {
	return &StorageError{Code: ErrorCode(code), Field: field, Operation: operation, MigrationVersion: migrationVersion, Reason: sanitizeReason(reason), Err: sanitizeCause(cause)}
}

// NewMigrationError constructs a typed migration failure error.
func NewMigrationError(version string, operation string, reason string, cause error) *StorageError {
	return NewStorageError(string(ErrorCodeMigrationFailed), "migration", operation, version, reason, cause)
}

func sanitizeReason(reason string) string {
	if reason == "" {
		return "redacted"
	}
	return RedactSensitive(reason)
}

func sanitizeCause(cause error) error {
	if cause == nil {
		return nil
	}
	return fmt.Errorf("%s", RedactSensitive(cause.Error()))
}
