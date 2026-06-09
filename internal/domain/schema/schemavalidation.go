package schema

import "encoding/json"

// SchemaIssueCode identifies a stable machine-readable issue category for schema domain validation.
type SchemaIssueCode string

const (
	// SchemaIssueCodeValidationFailed reports a generic schema field validation failure.
	SchemaIssueCodeValidationFailed SchemaIssueCode = "VALIDATION_FAILED"

	// SchemaIssueCodeRequired reports that a required schema domain field is missing or blank.
	SchemaIssueCodeRequired SchemaIssueCode = "SCHEMA_REQUIRED"

	// SchemaIssueCodeInvalidID reports an invalid primary key or parent reference identifier.
	SchemaIssueCodeInvalidID SchemaIssueCode = "SCHEMA_INVALID_ID"

	// SchemaIssueCodeInvalidName reports an invalid catalog or schema name value.
	SchemaIssueCodeInvalidName SchemaIssueCode = "SCHEMA_INVALID_NAME"

	// SchemaIssueCodeInvalidTime reports an invalid schema domain time value or audit-time ordering conflict.
	SchemaIssueCodeInvalidTime SchemaIssueCode = "SCHEMA_INVALID_TIME"

	// SchemaIssueCodeInvalidIdentity reports an incomplete or inconsistent schema identity value object.
	SchemaIssueCodeInvalidIdentity SchemaIssueCode = "SCHEMA_INVALID_IDENTITY"
)

// IsKnown reports whether the schema issue code belongs to the stable schema domain issue-code set.
func (c SchemaIssueCode) IsKnown() bool {
	switch c {
	case SchemaIssueCodeValidationFailed,
		SchemaIssueCodeRequired,
		SchemaIssueCodeInvalidID,
		SchemaIssueCodeInvalidName,
		SchemaIssueCodeInvalidTime,
		SchemaIssueCodeInvalidIdentity:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the schema issue code is outside the stable schema domain issue-code set.
func (c SchemaIssueCode) IsUnknown() bool {
	return !c.IsKnown()
}

// String returns the stable string representation used for persistence and transport.
// Unknown values are returned unchanged so callers can validate and report them explicitly.
func (c SchemaIssueCode) String() string {
	return string(c)
}

// MarshalJSON serializes the schema issue code as its stable string value.
// Unknown values are preserved so validation can classify them explicitly.
func (c SchemaIssueCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON restores the schema issue code from its serialized string value.
// Unknown strings are preserved instead of being coerced, allowing validation to reject them explicitly.
func (c *SchemaIssueCode) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*c = SchemaIssueCode(value)
	return nil
}

// SchemaIssueSeverity classifies how callers should present or handle a schema validation issue.
type SchemaIssueSeverity string

const (
	// SchemaIssueSeverityInfo marks a non-blocking informational schema issue.
	SchemaIssueSeverityInfo SchemaIssueSeverity = "info"

	// SchemaIssueSeverityWarning marks a schema issue callers may present without blocking acceptance.
	SchemaIssueSeverityWarning SchemaIssueSeverity = "warning"

	// SchemaIssueSeverityError marks a schema issue that blocks accepting or saving the model.
	SchemaIssueSeverityError SchemaIssueSeverity = "error"
)

// IsKnown reports whether the schema issue severity belongs to the stable schema severity set.
func (s SchemaIssueSeverity) IsKnown() bool {
	switch s {
	case SchemaIssueSeverityInfo,
		SchemaIssueSeverityWarning,
		SchemaIssueSeverityError:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the schema issue severity is outside the stable schema severity set.
func (s SchemaIssueSeverity) IsUnknown() bool {
	return !s.IsKnown()
}

// String returns the stable string representation used for persistence and transport.
// Unknown values are returned unchanged so callers can validate and report them explicitly.
func (s SchemaIssueSeverity) String() string {
	return string(s)
}

// MarshalJSON serializes the schema issue severity as its stable string value.
// Unknown values are preserved so validation can classify them explicitly.
func (s SchemaIssueSeverity) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON restores the schema issue severity from its serialized string value.
// Unknown strings are preserved instead of being coerced, allowing validation to reject them explicitly.
func (s *SchemaIssueSeverity) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*s = SchemaIssueSeverity(value)
	return nil
}

// SchemaValidationMode identifies the validation contract for draft objects or persisted snapshots.
type SchemaValidationMode string

const (
	// SchemaValidationModeDraft validates a new domain object before it is persisted.
	SchemaValidationModeDraft SchemaValidationMode = "draft"

	// SchemaValidationModePersisted validates an object that has been loaded from or prepared as a persisted snapshot.
	SchemaValidationModePersisted SchemaValidationMode = "persisted"
)

// IsKnown reports whether the validation mode belongs to the stable schema validation mode set.
func (m SchemaValidationMode) IsKnown() bool {
	switch m {
	case SchemaValidationModeDraft,
		SchemaValidationModePersisted:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the validation mode is outside the stable schema validation mode set.
func (m SchemaValidationMode) IsUnknown() bool {
	return !m.IsKnown()
}

// String returns the stable string representation used for persistence and transport.
// Unknown values are returned unchanged so callers can validate and report them explicitly.
func (m SchemaValidationMode) String() string {
	return string(m)
}

// MarshalJSON serializes the validation mode as its stable string value.
// Unknown values are preserved so validation can classify them explicitly.
func (m SchemaValidationMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

// UnmarshalJSON restores the validation mode from its serialized string value.
// Unknown strings are preserved instead of being coerced, allowing validation to reject them explicitly.
func (m *SchemaValidationMode) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*m = SchemaValidationMode(value)
	return nil
}

// SchemaValidationIssue is a field-level schema domain problem compatible with ConfigIssue and ApiIssue shapes.
type SchemaValidationIssue struct {
	// Path is the lower camelCase dotted field path, such as catalogName or identity.schemaName.
	Path string `json:"path"`

	// Code is the stable machine-readable schema issue category.
	Code SchemaIssueCode `json:"code"`

	// Severity tells callers whether the issue is informational, warning, or blocking.
	Severity SchemaIssueSeverity `json:"severity"`

	// Message is a safe user-readable explanation that must not contain credentials, user SQL, or generated data.
	Message string `json:"message"`
}
