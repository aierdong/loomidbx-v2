package schema

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"
)

// SchemaJSONError reports JSON decoding failures as schema-domain field-level issues.
type SchemaJSONError struct {
	// issues stores the field-level schema validation issues found during JSON decoding.
	issues []SchemaValidationIssue
}

// Error returns a compact diagnostic string for JSON decoding failures.
func (e SchemaJSONError) Error() string {
	if len(e.issues) == 0 {
		return "schema JSON decoding failed"
	}
	issue := e.issues[0]
	return fmt.Sprintf("schema JSON decoding failed at %s: %s", issue.Path, issue.Message)
}

// Issues returns the field-level schema validation issues detected while decoding JSON.
func (e SchemaJSONError) Issues() []SchemaValidationIssue {
	return append([]SchemaValidationIssue(nil), e.issues...)
}

// ValidateCatalog validates a catalog model using the explicit draft or persisted schema validation mode.
func ValidateCatalog(catalog DbCatalog, mode SchemaValidationMode) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 6)
	if mode.IsUnknown() {
		return append(issues, schemaValidationIssue("mode", SchemaIssueCodeValidationFailed, "validation mode must be draft or persisted"))
	}

	if mode == SchemaValidationModePersisted {
		if catalog.ID <= 0 {
			issues = append(issues, schemaValidationIssue("id", SchemaIssueCodeInvalidID, "id must be greater than zero for persisted catalog snapshots"))
		}
	} else if catalog.ID < 0 {
		issues = append(issues, schemaValidationIssue("id", SchemaIssueCodeInvalidID, "id must not be negative for draft catalogs"))
	}

	if catalog.ConnectionID <= 0 {
		issues = append(issues, schemaValidationIssue("connectionId", SchemaIssueCodeInvalidID, "connectionId must reference a saved connection"))
	}
	issues = append(issues, validateRequiredName("catalogName", catalog.CatalogName)...)
	issues = append(issues, validateOptionalTime("scannedAt", catalog.ScannedAt)...)
	issues = append(issues, validateAuditTimes(mode, catalog.CreatedAt, catalog.UpdatedAt)...)

	return issues
}

// ValidateSchema validates a schema model using the explicit draft or persisted schema validation mode.
func ValidateSchema(schema DbSchema, mode SchemaValidationMode) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 6)
	if mode.IsUnknown() {
		return append(issues, schemaValidationIssue("mode", SchemaIssueCodeValidationFailed, "validation mode must be draft or persisted"))
	}

	if mode == SchemaValidationModePersisted {
		if schema.ID <= 0 {
			issues = append(issues, schemaValidationIssue("id", SchemaIssueCodeInvalidID, "id must be greater than zero for persisted schema snapshots"))
		}
	} else if schema.ID < 0 {
		issues = append(issues, schemaValidationIssue("id", SchemaIssueCodeInvalidID, "id must not be negative for draft schemas"))
	}

	if schema.CatalogID <= 0 {
		issues = append(issues, schemaValidationIssue("catalogId", SchemaIssueCodeInvalidID, "catalogId must reference a saved catalog"))
	}
	issues = append(issues, validateImplicitAllowedName("schemaName", schema.SchemaName)...)
	issues = append(issues, validateOptionalTime("scannedAt", schema.ScannedAt)...)
	issues = append(issues, validateAuditTimes(mode, schema.CreatedAt, schema.UpdatedAt)...)

	return issues
}

// ValidateIdentity validates a schema identity value object without applying draft or persisted snapshot rules.
func ValidateIdentity(identity SchemaIdentity) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 3)
	if identity.ConnectionID <= 0 {
		issues = append(issues, schemaValidationIssue("connectionId", SchemaIssueCodeInvalidIdentity, "connectionId must reference a saved connection"))
	}
	issues = append(issues, validateRequiredName("catalogName", identity.CatalogName)...)
	issues = append(issues, validateImplicitAllowedName("schemaName", identity.SchemaName)...)
	return issues
}

// ValidateCatalogUniqueness expresses the in-memory uniqueness contract for catalogs without querying a database.
func ValidateCatalogUniqueness(catalogs []DbCatalog) []SchemaValidationIssue {
	seen := make(map[string]int, len(catalogs))
	issues := make([]SchemaValidationIssue, 0)
	for index, catalog := range catalogs {
		key := fmt.Sprintf("%d\x00%s", catalog.ConnectionID, strings.TrimSpace(catalog.CatalogName))
		if firstIndex, ok := seen[key]; ok {
			issues = append(issues, schemaValidationIssue("catalogName", SchemaIssueCodeValidationFailed, fmt.Sprintf("catalogName must be unique within the same connection; duplicate entries at positions %d and %d", firstIndex, index)))
			continue
		}
		seen[key] = index
	}
	return issues
}

// ValidateSchemaUniqueness expresses the in-memory uniqueness contract for schemas without querying a database.
func ValidateSchemaUniqueness(schemas []DbSchema) []SchemaValidationIssue {
	seen := make(map[string]int, len(schemas))
	issues := make([]SchemaValidationIssue, 0)
	for index, schema := range schemas {
		key := fmt.Sprintf("%d\x00%s", schema.CatalogID, schema.SchemaName)
		if firstIndex, ok := seen[key]; ok {
			issues = append(issues, schemaValidationIssue("schemaName", SchemaIssueCodeValidationFailed, fmt.Sprintf("schemaName must be unique within the same catalog; duplicate entries at positions %d and %d", firstIndex, index)))
			continue
		}
		seen[key] = index
	}
	return issues
}

func requireSchemaNameJSON(data []byte, path string) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	schemaName, ok := raw["schemaName"]
	if !ok || string(schemaName) == "null" {
		return schemaNameRequiredJSONError(path)
	}
	return nil
}

func schemaNameRequiredJSONError(path string) SchemaJSONError {
	return SchemaJSONError{issues: []SchemaValidationIssue{{
		Path:     path,
		Code:     SchemaIssueCodeRequired,
		Severity: SchemaIssueSeverityError,
		Message:  path + " is required and must be a string; use an empty string for an implicit schema",
	}}}
}

func schemaValidationIssue(path string, code SchemaIssueCode, message string) SchemaValidationIssue {
	return SchemaValidationIssue{
		Path:     path,
		Code:     code,
		Severity: SchemaIssueSeverityError,
		Message:  message,
	}
}

func validateRequiredName(path string, value string) []SchemaValidationIssue {
	if strings.TrimSpace(value) == "" {
		return []SchemaValidationIssue{schemaValidationIssue(path, SchemaIssueCodeRequired, path+" is required")}
	}
	return validateName(path, value)
}

func validateImplicitAllowedName(path string, value string) []SchemaValidationIssue {
	if value == "" {
		return nil
	}
	return validateName(path, value)
}

func validateName(path string, value string) []SchemaValidationIssue {
	if strings.TrimSpace(value) == "" {
		return []SchemaValidationIssue{schemaValidationIssue(path, SchemaIssueCodeInvalidName, path+" must not be blank; use an empty string only for an implicit schema")}
	}
	for _, r := range value {
		if r == '/' || r == '\\' || unicode.IsControl(r) {
			return []SchemaValidationIssue{schemaValidationIssue(path, SchemaIssueCodeInvalidName, path+" must not contain path separators or control characters")}
		}
	}
	return nil
}

func validateOptionalTime(path string, value *time.Time) []SchemaValidationIssue {
	if value != nil && value.IsZero() {
		return []SchemaValidationIssue{schemaValidationIssue(path, SchemaIssueCodeInvalidTime, path+" must be nil or a non-zero time")}
	}
	return nil
}

func validateAuditTimes(mode SchemaValidationMode, createdAt time.Time, updatedAt time.Time) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 2)
	if mode == SchemaValidationModePersisted {
		if createdAt.IsZero() {
			issues = append(issues, schemaValidationIssue("createdAt", SchemaIssueCodeInvalidTime, "createdAt is required for persisted snapshots"))
		}
		if updatedAt.IsZero() {
			issues = append(issues, schemaValidationIssue("updatedAt", SchemaIssueCodeInvalidTime, "updatedAt is required for persisted snapshots"))
		}
	}
	if !createdAt.IsZero() && !updatedAt.IsZero() && updatedAt.Before(createdAt) {
		issues = append(issues, schemaValidationIssue("updatedAt", SchemaIssueCodeInvalidTime, "updatedAt must not be earlier than createdAt"))
	}
	return issues
}
