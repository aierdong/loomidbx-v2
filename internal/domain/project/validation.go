package project

// ProjectIssueCode identifies a stable machine-readable Project validation issue category.
type ProjectIssueCode string

const (
	// ProjectIssueCodeValidationFailed reports that an issue object itself violates the validation issue contract.
	ProjectIssueCodeValidationFailed ProjectIssueCode = "VALIDATION_FAILED"

	// ProjectIssueCodeRequired reports that a required Project field is missing or blank.
	ProjectIssueCodeRequired ProjectIssueCode = "REQUIRED"

	// ProjectIssueCodeInvalidID reports an invalid Project, table, relation, or parent reference identifier.
	ProjectIssueCodeInvalidID ProjectIssueCode = "INVALID_ID"

	// ProjectIssueCodeInvalidRange reports an invalid Project numeric range such as row count or multiplier bounds.
	ProjectIssueCodeInvalidRange ProjectIssueCode = "INVALID_RANGE"

	// ProjectIssueCodeInvalidEnum reports an unknown Project enum value.
	ProjectIssueCodeInvalidEnum ProjectIssueCode = "INVALID_ENUM"

	// ProjectIssueCodeInvalidTime reports a missing or inconsistent Project audit timestamp.
	ProjectIssueCodeInvalidTime ProjectIssueCode = "INVALID_TIME"

	// ProjectIssueCodeDuplicateTable reports a duplicate table reference within one Project validation set.
	ProjectIssueCodeDuplicateTable ProjectIssueCode = "DUPLICATE_TABLE"

	// ProjectIssueCodeSQLRequired reports that relSourceSql is required for the selected relation value source.
	ProjectIssueCodeSQLRequired ProjectIssueCode = "SQL_REQUIRED"

	// ProjectIssueCodeParentRequired reports that parentProjectTableId is required for the selected relation value source.
	ProjectIssueCodeParentRequired ProjectIssueCode = "PARENT_REQUIRED"

	// ProjectIssueCodeOutOfScope reports boundary protection failures in tests or validation harnesses.
	ProjectIssueCodeOutOfScope ProjectIssueCode = "OUT_OF_SCOPE"
)

// IsKnown reports whether code belongs to the stable Project validation issue-code set.
func (code ProjectIssueCode) IsKnown() bool {
	switch code {
	case ProjectIssueCodeValidationFailed,
		ProjectIssueCodeRequired,
		ProjectIssueCodeInvalidID,
		ProjectIssueCodeInvalidRange,
		ProjectIssueCodeInvalidEnum,
		ProjectIssueCodeInvalidTime,
		ProjectIssueCodeDuplicateTable,
		ProjectIssueCodeSQLRequired,
		ProjectIssueCodeParentRequired,
		ProjectIssueCodeOutOfScope:
		return true
	default:
		return false
	}
}

// ProjectIssueSeverity classifies how callers should present or handle a Project validation issue.
type ProjectIssueSeverity string

const (
	// ProjectIssueSeverityInfo marks a non-blocking informational Project validation issue.
	ProjectIssueSeverityInfo ProjectIssueSeverity = "info"

	// ProjectIssueSeverityWarning marks a Project validation issue callers may present without blocking acceptance.
	ProjectIssueSeverityWarning ProjectIssueSeverity = "warning"

	// ProjectIssueSeverityError marks a Project validation issue that blocks accepting or saving the model.
	ProjectIssueSeverityError ProjectIssueSeverity = "error"
)

// IsKnown reports whether severity belongs to the stable Project validation issue severity set.
func (severity ProjectIssueSeverity) IsKnown() bool {
	switch severity {
	case ProjectIssueSeverityInfo, ProjectIssueSeverityWarning, ProjectIssueSeverityError:
		return true
	default:
		return false
	}
}

// ProjectValidationIssue is a JSON-compatible field-level Project validation problem.
type ProjectValidationIssue struct {
	// Path is the lower camelCase dotted field path, such as project.name or relations[0].relSourceSql.
	Path string `json:"path"`

	// Code is the stable machine-readable Project validation issue category.
	Code ProjectIssueCode `json:"code"`

	// Severity tells callers whether the issue is informational, warning, or blocking.
	Severity ProjectIssueSeverity `json:"severity"`

	// Message is a safe user-readable explanation that must not contain credentials, user SQL, or generated data.
	Message string `json:"message"`
}

// ProjectValidationIssues is a JSON-compatible collection of Project validation issues returned together.
type ProjectValidationIssues []ProjectValidationIssue

// NewProjectValidationIssue builds a blocking Project validation issue with a safe message for the given path and code.
func NewProjectValidationIssue(path string, code ProjectIssueCode) ProjectValidationIssue {
	return ProjectValidationIssue{
		Path:     path,
		Code:     code,
		Severity: ProjectIssueSeverityError,
		Message:  projectValidationMessage(code, path),
	}
}

func projectValidationMessage(code ProjectIssueCode, path string) string {
	switch code {
	case ProjectIssueCodeRequired:
		return fieldOrValue(path, "field") + " is required"
	case ProjectIssueCodeInvalidID:
		return fieldOrValue(path, "id") + " must reference a valid identity"
	case ProjectIssueCodeInvalidRange:
		return fieldOrValue(path, "range") + " must be within the allowed range"
	case ProjectIssueCodeInvalidEnum:
		return fieldOrValue(path, "enum") + " must be one of the stable values"
	case ProjectIssueCodeInvalidTime:
		return fieldOrValue(path, "time") + " must satisfy the audit time rules"
	case ProjectIssueCodeDuplicateTable:
		return "tableId must be unique within the Project"
	case ProjectIssueCodeSQLRequired:
		return "relSourceSql is required for the chosen relation value source"
	case ProjectIssueCodeParentRequired:
		return "parentProjectTableId is required for the selected relation value source"
	case ProjectIssueCodeOutOfScope:
		return "Project validation must stay within the domain boundary"
	default:
		return "Project validation issue is invalid"
	}
}

func fieldOrValue(path string, fallback string) string {
	if path == "" {
		return fallback
	}
	return path
}
