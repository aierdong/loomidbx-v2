package connection

// ValidationCode is a stable machine-readable code for a connection validation issue.
type ValidationCode string

const (
	// ValidationCodeRequired reports that a required field is missing or blank.
	ValidationCodeRequired ValidationCode = "required"

	// ValidationCodeUnknownDatabaseType reports that a database type is outside the known connection type set.
	ValidationCodeUnknownDatabaseType ValidationCode = "unknown_database_type"

	// ValidationCodeInvalidPort reports that a port is outside the valid TCP port range.
	ValidationCodeInvalidPort ValidationCode = "invalid_port"

	// ValidationCodeSensitiveParamNotAllowed reports that a sensitive extension parameter must use the credential boundary.
	ValidationCodeSensitiveParamNotAllowed ValidationCode = "sensitive_param_not_allowed"
)

// ValidationError describes one field-level validation issue using a safe message.
type ValidationError struct {
	// Field stores the stable field path that failed validation.
	Field string `json:"field"`

	// Code stores the machine-readable validation error code.
	Code ValidationCode `json:"code"`

	// Message stores a non-sensitive diagnostic message suitable for callers and UI surfaces.
	Message string `json:"message"`
}

// ValidationResult stores all validation issues detected for a connection configuration.
type ValidationResult struct {
	// Errors stores the field-level validation errors found during validation.
	Errors []ValidationError `json:"errors,omitempty"`
}
