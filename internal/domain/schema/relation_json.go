package schema

import "encoding/json"

// DecodeForeignKeyJSON decodes a foreign key JSON payload and returns field-level issues instead of exposing transport errors.
func DecodeForeignKeyJSON(data []byte, mode SchemaValidationMode) (ForeignKey, []SchemaValidationIssue) {
	var foreignKey ForeignKey
	if err := json.Unmarshal(data, &foreignKey); err != nil {
		return foreignKey, []SchemaValidationIssue{schemaValidationIssue("payload", SchemaIssueCodeValidationFailed, "foreign key JSON payload must be a valid object")}
	}
	return foreignKey, ValidateForeignKey(foreignKey, mode)
}

// DecodeTableRelationJSON decodes a table relation JSON payload and returns field-level issues instead of exposing transport errors.
func DecodeTableRelationJSON(data []byte, mode SchemaValidationMode) (TableRelation, []SchemaValidationIssue) {
	var relation TableRelation
	if err := json.Unmarshal(data, &relation); err != nil {
		return relation, []SchemaValidationIssue{schemaValidationIssue("payload", SchemaIssueCodeValidationFailed, "table relation JSON payload must be a valid object")}
	}
	return relation, ValidateTableRelation(relation, mode)
}
