package schema

import (
	"encoding/json"
	"fmt"
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
