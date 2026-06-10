package project

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"
)

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

// ValidateProject validates a Project aggregate root using only pure in-memory Project domain rules.
func ValidateProject(project Project, persisted bool) []ProjectValidationIssue {
	issues := make([]ProjectValidationIssue, 0, 6)
	issues = append(issues, validateIdentityByMode("project.id", project.ID, persisted)...)
	if project.ConnectionID <= 0 {
		issues = append(issues, NewProjectValidationIssue("project.connectionId", ProjectIssueCodeInvalidID))
	}
	issues = append(issues, validateRequiredProjectString("project.name", project.Name, 100)...)
	issues = append(issues, validateOptionalProjectString("project.description", project.Description)...)
	issues = append(issues, validateProjectAuditTimes("project", project.CreatedAt, project.UpdatedAt, persisted)...)
	return issues
}

// ValidateProjectTable validates a ProjectTable using only pure in-memory Project domain rules.
func ValidateProjectTable(table ProjectTable, persisted bool) []ProjectValidationIssue {
	issues := make([]ProjectValidationIssue, 0, 7)
	issues = append(issues, validateIdentityByMode("projectTable.id", table.ID, persisted)...)
	if table.ProjectID <= 0 {
		issues = append(issues, NewProjectValidationIssue("projectTable.projectId", ProjectIssueCodeInvalidID))
	}
	if table.TableID <= 0 {
		issues = append(issues, NewProjectValidationIssue("projectTable.tableId", ProjectIssueCodeInvalidID))
	}
	if table.RowCount != nil && *table.RowCount < 0 {
		issues = append(issues, NewProjectValidationIssue("projectTable.rowCount", ProjectIssueCodeInvalidRange))
	}
	if table.ExecutionOrder <= 0 {
		issues = append(issues, NewProjectValidationIssue("projectTable.executionOrder", ProjectIssueCodeInvalidRange))
	}
	issues = append(issues, validateProjectAuditTimes("projectTable", table.CreatedAt, table.UpdatedAt, persisted)...)
	return issues
}

// ValidateProjectTables validates ProjectTable entries and duplicate table references within the same Project.
func ValidateProjectTables(tables []ProjectTable, persisted bool) []ProjectValidationIssue {
	issues := make([]ProjectValidationIssue, 0, len(tables))
	seen := make(map[string]int, len(tables))
	for index, table := range tables {
		issues = append(issues, prefixProjectValidationIssues(ValidateProjectTable(table, persisted), fmt.Sprintf("tables[%d]", index), "projectTable")...)
		if table.ProjectID <= 0 || table.TableID <= 0 {
			continue
		}
		key := fmt.Sprintf("%d\x00%d", table.ProjectID, table.TableID)
		if _, ok := seen[key]; ok {
			issues = append(issues, NewProjectValidationIssue(fmt.Sprintf("tables[%d].tableId", index), ProjectIssueCodeDuplicateTable))
			continue
		}
		seen[key] = index
	}
	return issues
}

// ValidateProjectTableRelation validates a ProjectTableRelation using only pure in-memory Project domain rules.
func ValidateProjectTableRelation(relation ProjectTableRelation, persisted bool) []ProjectValidationIssue {
	issues := make([]ProjectValidationIssue, 0, 10)
	issues = append(issues, validateIdentityByMode("projectTableRelation.id", relation.ID, persisted)...)
	if relation.ProjectID <= 0 {
		issues = append(issues, NewProjectValidationIssue("projectTableRelation.projectId", ProjectIssueCodeInvalidID))
	}
	if relation.TableRelationID <= 0 {
		issues = append(issues, NewProjectValidationIssue("projectTableRelation.tableRelationId", ProjectIssueCodeInvalidID))
	}
	if relation.ParentProjectTableID != nil && *relation.ParentProjectTableID <= 0 {
		issues = append(issues, NewProjectValidationIssue("projectTableRelation.parentProjectTableId", ProjectIssueCodeInvalidID))
	}
	if relation.ChildProjectTableID <= 0 {
		issues = append(issues, NewProjectValidationIssue("projectTableRelation.childProjectTableId", ProjectIssueCodeInvalidID))
	}
	if relation.MultiplierMin < 0 {
		issues = append(issues, NewProjectValidationIssue("projectTableRelation.multiplierMin", ProjectIssueCodeInvalidRange))
	} else if relation.MultiplierMax < relation.MultiplierMin {
		issues = append(issues, NewProjectValidationIssue("projectTableRelation.multiplierMax", ProjectIssueCodeInvalidRange))
	}
	issues = append(issues, validateRelationValueSourceCombination(relation)...)
	issues = append(issues, validateProjectAuditTimes("projectTableRelation", relation.CreatedAt, relation.UpdatedAt, persisted)...)
	return issues
}

// DecodeProjectTableJSON decodes a ProjectTable JSON payload, reports required-field presence issues, and then applies ProjectTable validation.
func DecodeProjectTableJSON(data []byte, persisted bool) (ProjectTable, []ProjectValidationIssue, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return ProjectTable{}, nil, err
	}

	var table ProjectTable
	if err := json.Unmarshal(data, &table); err != nil {
		return ProjectTable{}, nil, err
	}

	requiredFields := []string{"id", "projectId", "tableId", "rowCount", "truncateBefore", "executionOrder"}
	if persisted {
		requiredFields = append(requiredFields, "createdAt", "updatedAt")
	}
	missing := missingProjectJSONFields(fields, requiredFields...)
	issues := make([]ProjectValidationIssue, 0, len(missing))
	issues = append(issues, requireMissingProjectJSONFields("projectTable", missing, requiredFields...)...)
	issues = append(issues, filterProjectValidationIssuesForMissingFields(ValidateProjectTable(table, persisted), "projectTable", missing)...)
	return table, issues, nil
}

// DecodeProjectTableRelationJSON decodes a ProjectTableRelation JSON payload, reports required-field presence issues, and then applies ProjectTableRelation validation.
func DecodeProjectTableRelationJSON(data []byte, persisted bool) (ProjectTableRelation, []ProjectValidationIssue, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return ProjectTableRelation{}, nil, err
	}

	var relation ProjectTableRelation
	if err := json.Unmarshal(data, &relation); err != nil {
		return ProjectTableRelation{}, nil, err
	}

	requiredFields := []string{"id", "projectId", "tableRelationId", "parentProjectTableId", "childProjectTableId", "multiplierMin", "multiplierMax", "relValueSource"}
	if persisted {
		requiredFields = append(requiredFields, "createdAt", "updatedAt")
	}
	missing := missingProjectJSONFields(fields, requiredFields...)
	issues := make([]ProjectValidationIssue, 0, len(missing))
	issues = append(issues, requireMissingProjectJSONFields("projectTableRelation", missing, requiredFields...)...)
	issues = append(issues, filterProjectValidationIssuesForMissingFields(ValidateProjectTableRelation(relation, persisted), "projectTableRelation", missing)...)
	return relation, issues, nil
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

func validateIdentityByMode(path string, id int64, persisted bool) []ProjectValidationIssue {
	if id < 0 || persisted && id <= 0 {
		return []ProjectValidationIssue{NewProjectValidationIssue(path, ProjectIssueCodeInvalidID)}
	}
	return nil
}

func validateRequiredProjectString(path string, value string, maxRunes int) []ProjectValidationIssue {
	if strings.TrimSpace(value) == "" {
		return []ProjectValidationIssue{NewProjectValidationIssue(path, ProjectIssueCodeRequired)}
	}
	if maxRunes > 0 && len([]rune(value)) > maxRunes || containsProjectControlRune(value) {
		return []ProjectValidationIssue{NewProjectValidationIssue(path, ProjectIssueCodeInvalidRange)}
	}
	return nil
}

func validateOptionalProjectString(path string, value string) []ProjectValidationIssue {
	if value == "" {
		return nil
	}
	if containsProjectControlRune(value) {
		return []ProjectValidationIssue{NewProjectValidationIssue(path, ProjectIssueCodeInvalidRange)}
	}
	return nil
}

func containsProjectControlRune(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func validateProjectAuditTimes(prefix string, createdAt time.Time, updatedAt time.Time, persisted bool) []ProjectValidationIssue {
	issues := make([]ProjectValidationIssue, 0, 2)
	if persisted {
		if createdAt.IsZero() {
			issues = append(issues, NewProjectValidationIssue(prefix+".createdAt", ProjectIssueCodeInvalidTime))
		}
		if updatedAt.IsZero() {
			issues = append(issues, NewProjectValidationIssue(prefix+".updatedAt", ProjectIssueCodeInvalidTime))
		}
	}
	if !createdAt.IsZero() && !updatedAt.IsZero() && updatedAt.Before(createdAt) {
		issues = append(issues, NewProjectValidationIssue(prefix+".updatedAt", ProjectIssueCodeInvalidTime))
	}
	return issues
}

func validateRelationValueSourceCombination(relation ProjectTableRelation) []ProjectValidationIssue {
	issues := make([]ProjectValidationIssue, 0, 2)
	if relation.RelValueSource == "" {
		return append(issues, NewProjectValidationIssue("projectTableRelation.relValueSource", ProjectIssueCodeRequired))
	}
	if !relation.RelValueSource.IsKnown() {
		return append(issues, NewProjectValidationIssue("projectTableRelation.relValueSource", ProjectIssueCodeInvalidEnum))
	}
	parentMissing := relation.ParentProjectTableID == nil
	sqlMissing := strings.TrimSpace(relation.RelSourceSQL) == ""
	switch relation.RelValueSource {
	case RelationValueSourceFromExecution:
		if parentMissing {
			issues = append(issues, NewProjectValidationIssue("projectTableRelation.parentProjectTableId", ProjectIssueCodeParentRequired))
		}
	case RelationValueSourceFromDBQuery:
		if sqlMissing {
			issues = append(issues, NewProjectValidationIssue("projectTableRelation.relSourceSql", ProjectIssueCodeSQLRequired))
		}
	case RelationValueSourceMerged:
		if parentMissing {
			issues = append(issues, NewProjectValidationIssue("projectTableRelation.parentProjectTableId", ProjectIssueCodeParentRequired))
		}
		if sqlMissing {
			issues = append(issues, NewProjectValidationIssue("projectTableRelation.relSourceSql", ProjectIssueCodeSQLRequired))
		}
	}
	return issues
}

func requireMissingProjectJSONFields(prefix string, missing map[string]bool, names ...string) []ProjectValidationIssue {
	issues := make([]ProjectValidationIssue, 0, len(names))
	for _, name := range names {
		if missing[name] {
			issues = append(issues, NewProjectValidationIssue(prefix+"."+name, ProjectIssueCodeRequired))
		}
	}
	return issues
}

func missingProjectJSONFields(fields map[string]json.RawMessage, names ...string) map[string]bool {
	missing := make(map[string]bool)
	for _, name := range names {
		if _, ok := fields[name]; !ok {
			missing[name] = true
		}
	}
	return missing
}

func filterProjectValidationIssuesForMissingFields(issues []ProjectValidationIssue, prefix string, missing map[string]bool) []ProjectValidationIssue {
	if len(missing) == 0 {
		return issues
	}
	filtered := issues[:0]
	for _, issue := range issues {
		if field := strings.TrimPrefix(issue.Path, prefix+"."); missing[field] {
			continue
		}
		filtered = append(filtered, issue)
	}
	return filtered
}

func prefixProjectValidationIssues(issues []ProjectValidationIssue, prefix string, oldPrefix string) []ProjectValidationIssue {
	prefixed := make([]ProjectValidationIssue, 0, len(issues))
	for _, issue := range issues {
		if issue.Path == oldPrefix {
			issue.Path = prefix
		} else if strings.HasPrefix(issue.Path, oldPrefix+".") {
			issue.Path = prefix + strings.TrimPrefix(issue.Path, oldPrefix)
		}
		issue.Message = projectValidationMessage(issue.Code, issue.Path)
		prefixed = append(prefixed, issue)
	}
	return prefixed
}
