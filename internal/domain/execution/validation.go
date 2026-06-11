package execution

import (
	"encoding/json"
	"fmt"
	"strings"
)

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

// Validate checks the generation job aggregate and its nested execution records using only in-memory domain data.
// It returns all field-level issues found, including nested model, parent reference, and collection uniqueness problems.
func (j GenerationJob) Validate() ValidationResult {
	result := ValidationResult{}
	taskResult := j.Task.Validate()
	if taskResult.HasIssues() {
		result.AddIssue("task", ValidationIssueCodeInvalidNestedModel, "task must be a valid execution task")
		appendPrefixedIssues(&result, "task", taskResult.Issues)
	}

	seenTableIDs := map[int64]struct{}{}
	seenExecutionOrders := map[int]struct{}{}
	for index, tableResult := range j.TableResults {
		prefix := fmt.Sprintf("tableResults[%d]", index)
		tableValidation := tableResult.Validate()
		if tableValidation.HasIssues() {
			result.AddIssue(prefix, ValidationIssueCodeInvalidNestedModel, "table result must be valid")
			appendPrefixedIssues(&result, prefix, tableValidation.Issues)
		}

		if j.Task.ID > 0 && tableResult.ExecutionTaskID > 0 && tableResult.ExecutionTaskID != j.Task.ID {
			result.AddIssue(prefix+".executionTaskId", ValidationIssueCodeInvalidReference, "executionTaskId must reference the parent task")
		}

		if tableResult.TableID != nil && *tableResult.TableID > 0 {
			if _, exists := seenTableIDs[*tableResult.TableID]; exists {
				result.AddIssue(prefix+".tableId", ValidationIssueCodeInvalidReference, "tableId must be unique within the generation job")
			} else {
				seenTableIDs[*tableResult.TableID] = struct{}{}
			}
		}

		if tableResult.ExecutionOrder > 0 {
			if _, exists := seenExecutionOrders[tableResult.ExecutionOrder]; exists {
				result.AddIssue(prefix+".executionOrder", ValidationIssueCodeInvalidRange, "executionOrder must be unique within the generation job")
			} else {
				seenExecutionOrders[tableResult.ExecutionOrder] = struct{}{}
			}
		}
	}

	return result
}

// Validate checks the execution task master record for required fields, references, enum values, ranges, and time ordering.
// It performs no database, execution engine, service, or UI lookups and returns all field-level issues found.
func (t ExecutionTask) Validate() ValidationResult {
	result := ValidationResult{}
	if t.ID < 0 {
		result.AddIssue("id", ValidationIssueCodeInvalidRange, "id must be zero for new records or a positive value")
	}
	if t.ProjectID <= 0 {
		result.AddIssue("projectId", ValidationIssueCodeInvalidReference, "projectId must reference an existing project")
	}
	validateTrimmedString(&result, "taskName", t.TaskName, 200)
	if t.Status.IsUnknown() {
		result.AddIssue("status", ValidationIssueCodeInvalidEnum, "status must be a known execution task status")
	}
	if t.StartedAt.IsZero() {
		result.AddIssue("startedAt", ValidationIssueCodeRequired, "startedAt is required")
	}
	if t.EndedAt != nil && !t.StartedAt.IsZero() && t.EndedAt.Before(t.StartedAt) {
		result.AddIssue("endedAt", ValidationIssueCodeInvalidTimeRange, "endedAt must not be before startedAt")
	}
	if t.CreatedAt.IsZero() {
		result.AddIssue("createdAt", ValidationIssueCodeRequired, "createdAt is required")
	}

	return result
}

// Validate checks the table-level execution result for required fields, references, enum values, ranges, time ordering, and failed-result error snapshots.
// It performs only pure domain validation and returns all field-level issues found.
func (r ExecutionTableResult) Validate() ValidationResult {
	result := ValidationResult{}
	if r.ID < 0 {
		result.AddIssue("id", ValidationIssueCodeInvalidRange, "id must be zero for new records or a positive value")
	}
	if r.ExecutionTaskID <= 0 {
		result.AddIssue("executionTaskId", ValidationIssueCodeInvalidReference, "executionTaskId must reference an execution task")
	}
	if r.TableID != nil && *r.TableID <= 0 {
		result.AddIssue("tableId", ValidationIssueCodeInvalidReference, "tableId must reference a table when present")
	}
	validateTrimmedString(&result, "tableNameSnapshot", r.TableNameSnapshot, 255)
	validateTrimmedString(&result, "schemaNameSnapshot", r.SchemaNameSnapshot, 255)
	if r.RowsWritten < 0 {
		result.AddIssue("rowsWritten", ValidationIssueCodeInvalidRange, "rowsWritten must be greater than or equal to zero")
	}
	if r.Status.IsUnknown() {
		result.AddIssue("status", ValidationIssueCodeInvalidEnum, "status must be a known execution table status")
	}
	if r.ExecutionOrder < 1 {
		result.AddIssue("executionOrder", ValidationIssueCodeInvalidRange, "executionOrder must be greater than or equal to one")
	}
	if r.CreatedAt.IsZero() {
		result.AddIssue("createdAt", ValidationIssueCodeRequired, "createdAt is required")
	}
	if r.UpdatedAt.IsZero() {
		result.AddIssue("updatedAt", ValidationIssueCodeRequired, "updatedAt is required")
	}
	if !r.CreatedAt.IsZero() && !r.UpdatedAt.IsZero() && r.UpdatedAt.Before(r.CreatedAt) {
		result.AddIssue("updatedAt", ValidationIssueCodeInvalidTimeRange, "updatedAt must not be before createdAt")
	}
	if r.Status == ExecutionTableStatusFailed {
		if r.ErrorSnapshot == nil {
			result.AddIssue("errorSnapshot", ValidationIssueCodeRequired, "errorSnapshot is required for failed table results")
		} else {
			appendPrefixedIssues(&result, "errorSnapshot", r.ErrorSnapshot.Validate().Issues)
		}
	} else if r.ErrorSnapshot != nil {
		appendPrefixedIssues(&result, "errorSnapshot", r.ErrorSnapshot.Validate().Issues)
	}

	return result
}

// Validate checks the execution error snapshot for required safe diagnostic fields.
// It returns issues without echoing raw error text, credentials, user SQL, or generated data into diagnostic messages.
func (s ExecutionErrorSnapshot) Validate() ValidationResult {
	result := ValidationResult{}
	if strings.TrimSpace(s.Code) == "" {
		result.AddIssue("code", ValidationIssueCodeRequired, "code is required")
	}
	if strings.TrimSpace(s.Message) == "" {
		result.AddIssue("message", ValidationIssueCodeRequired, "message is required")
	} else if containsSensitiveMarker(s.Message) {
		result.AddIssue("message", ValidationIssueCodeSensitiveValueNotAllowed, "message must not contain sensitive values")
	}
	if s.OccurredAt.IsZero() {
		result.AddIssue("occurredAt", ValidationIssueCodeRequired, "occurredAt is required")
	}

	return result
}

func appendPrefixedIssues(result *ValidationResult, prefix string, issues []ValidationIssue) {
	for _, issue := range issues {
		path := prefix
		if issue.Path != "" {
			path += "." + issue.Path
		}
		result.AddIssue(path, issue.Code, issue.Message)
	}
}

func validateTrimmedString(result *ValidationResult, path string, value string, maxLength int) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		result.AddIssue(path, ValidationIssueCodeRequired, path+" is required")
		return
	}
	if len(trimmed) > maxLength {
		result.AddIssue(path, ValidationIssueCodeTooLong, path+" must not exceed its maximum length")
	}
}

func containsSensitiveMarker(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{
		"password",
		"credential",
		"token",
		"secret",
		"select *",
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
