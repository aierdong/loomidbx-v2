package execution

import "encoding/json"

// ValidationIssueCode identifies a stable machine-readable execution validation issue category.
type ValidationIssueCode string

const (
	// ValidationIssueCodeRequired reports that a required execution domain field is missing or blank.
	ValidationIssueCodeRequired ValidationIssueCode = "REQUIRED"

	// ValidationIssueCodeTooLong reports that an execution domain text field exceeds its maximum length.
	ValidationIssueCodeTooLong ValidationIssueCode = "TOO_LONG"

	// ValidationIssueCodeInvalidEnum reports that an execution domain enum value is outside the stable value set.
	ValidationIssueCodeInvalidEnum ValidationIssueCode = "INVALID_ENUM"

	// ValidationIssueCodeInvalidRange reports that an execution domain numeric value is outside the accepted range.
	ValidationIssueCodeInvalidRange ValidationIssueCode = "INVALID_RANGE"

	// ValidationIssueCodeInvalidReference reports that an execution domain parent or related identity reference is invalid.
	ValidationIssueCodeInvalidReference ValidationIssueCode = "INVALID_REFERENCE"

	// ValidationIssueCodeInvalidTimeRange reports that execution domain time values are missing or ordered incorrectly.
	ValidationIssueCodeInvalidTimeRange ValidationIssueCode = "INVALID_TIME_RANGE"

	// ValidationIssueCodeInvalidNestedModel reports that a nested execution domain model failed validation.
	ValidationIssueCodeInvalidNestedModel ValidationIssueCode = "INVALID_NESTED_MODEL"

	// ValidationIssueCodeSensitiveValueNotAllowed reports that an execution error message contains disallowed sensitive content.
	ValidationIssueCodeSensitiveValueNotAllowed ValidationIssueCode = "SENSITIVE_VALUE_NOT_ALLOWED"
)

// IsKnown reports whether the validation issue code belongs to the stable execution validation issue-code set.
func (c ValidationIssueCode) IsKnown() bool {
	switch c {
	case ValidationIssueCodeRequired,
		ValidationIssueCodeTooLong,
		ValidationIssueCodeInvalidEnum,
		ValidationIssueCodeInvalidRange,
		ValidationIssueCodeInvalidReference,
		ValidationIssueCodeInvalidTimeRange,
		ValidationIssueCodeInvalidNestedModel,
		ValidationIssueCodeSensitiveValueNotAllowed:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the validation issue code is outside the stable execution validation issue-code set.
func (c ValidationIssueCode) IsUnknown() bool {
	return !c.IsKnown()
}

// String returns the stable string representation used for persistence and transport.
// Unknown values are returned unchanged so callers can validate and report them explicitly.
func (c ValidationIssueCode) String() string {
	return string(c)
}

// MarshalJSON serializes the validation issue code as its stable string value.
// Unknown values are preserved so validation can classify them explicitly.
func (c ValidationIssueCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON restores the validation issue code from its serialized string value.
// Unknown strings are preserved instead of being coerced, allowing validation to reject them explicitly.
func (c *ValidationIssueCode) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*c = ValidationIssueCode(value)
	return nil
}

// ValidationIssue is a field-level execution domain validation problem with a safe diagnostic message.
type ValidationIssue struct {
	// Path is the JSON field path, such as projectId or tableResults[0].rowsWritten.
	Path string `json:"path"`

	// Code is the stable machine-readable execution validation issue category.
	Code ValidationIssueCode `json:"code"`

	// Message is a safe user-readable explanation that must not contain credentials, user SQL, or generated data.
	Message string `json:"message"`
}

// ValidationResult stores all field-level execution validation issues detected for one validation pass.
type ValidationResult struct {
	// Issues stores the field-level validation issues found during validation.
	Issues []ValidationIssue `json:"issues,omitempty"`
}

// AddIssue appends one field-level validation issue to the result.
func (r *ValidationResult) AddIssue(path string, code ValidationIssueCode, message string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Path:    path,
		Code:    code,
		Message: message,
	})
}

// HasIssues reports whether the result contains one or more validation issues.
func (r ValidationResult) HasIssues() bool {
	return len(r.Issues) > 0
}
