package schema

import "encoding/json"

var foreignKeyRequiredJSONFields = []string{"id", "tableId", "fkName", "referencedTableId", "columnIds", "referencedColumnIds", "createdAt"}
var tableRelationRequiredJSONFields = []string{"id", "relationType", "parentTableId", "childTableId", "parentColumnIds", "childColumnIds", "multiplierMin", "multiplierMax", "isLogical", "createdAt", "updatedAt"}

// DecodeForeignKeyJSON decodes a foreign key JSON payload and returns field-level issues instead of exposing transport errors.
func DecodeForeignKeyJSON(data []byte, mode SchemaValidationMode) (ForeignKey, []SchemaValidationIssue) {
	var foreignKey ForeignKey
	presenceIssues, ok := relationRequiredFieldIssues(data, foreignKeyRequiredJSONFields, "foreign key JSON payload must be a valid object")
	if !ok {
		return foreignKey, presenceIssues
	}
	if err := json.Unmarshal(data, &foreignKey); err != nil {
		return foreignKey, []SchemaValidationIssue{schemaValidationIssue("payload", SchemaIssueCodeValidationFailed, "foreign key JSON payload must be a valid object")}
	}
	return foreignKey, mergeRelationJSONIssues(presenceIssues, ValidateForeignKey(foreignKey, mode))
}

// DecodeTableRelationJSON decodes a table relation JSON payload and returns field-level issues instead of exposing transport errors.
func DecodeTableRelationJSON(data []byte, mode SchemaValidationMode) (TableRelation, []SchemaValidationIssue) {
	var relation TableRelation
	presenceIssues, ok := relationRequiredFieldIssues(data, tableRelationRequiredJSONFields, "table relation JSON payload must be a valid object")
	if !ok {
		return relation, presenceIssues
	}
	if err := json.Unmarshal(data, &relation); err != nil {
		return relation, []SchemaValidationIssue{schemaValidationIssue("payload", SchemaIssueCodeValidationFailed, "table relation JSON payload must be a valid object")}
	}
	return relation, mergeRelationJSONIssues(presenceIssues, ValidateTableRelation(relation, mode))
}

func relationRequiredFieldIssues(data []byte, fields []string, invalidMessage string) ([]SchemaValidationIssue, bool) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil || raw == nil {
		return []SchemaValidationIssue{schemaValidationIssue("payload", SchemaIssueCodeValidationFailed, invalidMessage)}, false
	}

	issues := make([]SchemaValidationIssue, 0, len(fields))
	for _, field := range fields {
		value, ok := raw[field]
		if !ok || string(value) == "null" {
			issues = append(issues, schemaValidationIssue(field, SchemaIssueCodeRequired, field+" is required"))
		}
	}
	return issues, true
}

func mergeRelationJSONIssues(presenceIssues []SchemaValidationIssue, validationIssues []SchemaValidationIssue) []SchemaValidationIssue {
	if len(presenceIssues) == 0 {
		return validationIssues
	}

	seenPresencePath := make(map[string]struct{}, len(presenceIssues))
	issues := make([]SchemaValidationIssue, 0, len(presenceIssues)+len(validationIssues))
	for _, issue := range presenceIssues {
		issues = append(issues, issue)
		seenPresencePath[issue.Path] = struct{}{}
	}
	for _, issue := range validationIssues {
		if _, exists := seenPresencePath[issue.Path]; exists {
			continue
		}
		issues = append(issues, issue)
	}
	return issues
}
