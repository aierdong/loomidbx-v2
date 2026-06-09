package schema

// SchemaIssueSeverity classifies how callers should present or handle a schema-domain issue.
type SchemaIssueSeverity string

const (
	// SchemaIssueSeverityInfo marks non-blocking schema-domain information.
	SchemaIssueSeverityInfo SchemaIssueSeverity = "info"

	// SchemaIssueSeverityWarning marks a schema-domain risk that callers may show without blocking acceptance.
	SchemaIssueSeverityWarning SchemaIssueSeverity = "warning"

	// SchemaIssueSeverityError marks a schema-domain issue that blocks accepting or saving the model.
	SchemaIssueSeverityError SchemaIssueSeverity = "error"
)

// SchemaIssueCode identifies the stable machine-readable category for a schema-domain validation issue.
type SchemaIssueCode string

const (
	// SchemaIssueCodeValidationFailed reports a generic field validation failure.
	SchemaIssueCodeValidationFailed SchemaIssueCode = "VALIDATION_FAILED"

	// SchemaIssueCodeRequired reports that a required schema-domain field is missing or blank.
	SchemaIssueCodeRequired SchemaIssueCode = "SCHEMA_REQUIRED"

	// SchemaIssueCodeInvalidID reports that an identity or parent reference is invalid.
	SchemaIssueCodeInvalidID SchemaIssueCode = "SCHEMA_INVALID_ID"

	// SchemaIssueCodeInvalidName reports that a catalog or schema name is invalid.
	SchemaIssueCodeInvalidName SchemaIssueCode = "SCHEMA_INVALID_NAME"

	// SchemaIssueCodeInvalidTime reports that an audit or scan time field is invalid.
	SchemaIssueCodeInvalidTime SchemaIssueCode = "SCHEMA_INVALID_TIME"

	// SchemaIssueCodeInvalidIdentity reports that a schema identity value object is incomplete or inconsistent.
	SchemaIssueCodeInvalidIdentity SchemaIssueCode = "SCHEMA_INVALID_IDENTITY"
)

// SchemaValidationMode selects the field rules for draft objects or persisted snapshots.
type SchemaValidationMode string

const (
	// SchemaValidationModeDraft validates a new domain object before storage assigns persisted fields.
	SchemaValidationModeDraft SchemaValidationMode = "draft"

	// SchemaValidationModePersisted validates a stored or snapshot object that must include persisted fields.
	SchemaValidationModePersisted SchemaValidationMode = "persisted"
)

// SchemaValidationIssue describes one field-level schema-domain issue using a safe message.
type SchemaValidationIssue struct {
	// Path stores the lower camelCase field path, such as catalogName or identity.schemaName.
	Path string `json:"path"`

	// Code stores the stable machine-readable schema issue code.
	Code SchemaIssueCode `json:"code"`

	// Severity stores whether this issue is informational, warning, or blocking.
	Severity SchemaIssueSeverity `json:"severity"`

	// Message stores a safe diagnostic message without credentials, user SQL, or generated data.
	Message string `json:"message"`
}
