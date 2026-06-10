package schema

import "encoding/json"

// DecodeTableJSON decodes a table JSON payload and returns field-level issues instead of exposing transport errors.
// Full presence diagnostics are implemented by later JSON tasks; this scaffold preserves the stable helper signature.
func DecodeTableJSON(data []byte, mode SchemaValidationMode) (DbTable, []SchemaValidationIssue) {
	var table DbTable
	if err := json.Unmarshal(data, &table); err != nil {
		return table, []SchemaValidationIssue{schemaValidationIssue("payload", SchemaIssueCodeValidationFailed, "table JSON payload must be a valid object")}
	}
	return table, ValidateTable(table, mode)
}

// DecodeColumnJSON decodes a column JSON payload and returns field-level issues instead of exposing transport errors.
// Full presence diagnostics are implemented by later JSON tasks; this scaffold preserves the stable helper signature.
func DecodeColumnJSON(data []byte, mode SchemaValidationMode) (DbColumn, []SchemaValidationIssue) {
	var column DbColumn
	if err := json.Unmarshal(data, &column); err != nil {
		return column, []SchemaValidationIssue{schemaValidationIssue("payload", SchemaIssueCodeValidationFailed, "column JSON payload must be a valid object")}
	}
	return column, ValidateColumn(column, mode)
}

// DecodeConstraintJSON decodes a constraint JSON payload and returns field-level issues instead of exposing transport errors.
// Full presence diagnostics are implemented by later JSON tasks; this scaffold preserves the stable helper signature.
func DecodeConstraintJSON(data []byte, mode SchemaValidationMode) (TableConstraint, []SchemaValidationIssue) {
	var constraint TableConstraint
	if err := json.Unmarshal(data, &constraint); err != nil {
		return constraint, []SchemaValidationIssue{schemaValidationIssue("payload", SchemaIssueCodeValidationFailed, "constraint JSON payload must be a valid object")}
	}
	return constraint, ValidateConstraint(constraint, mode)
}

// DecodeLogicalTypeJSON decodes a logical type JSON payload and returns field-level issues instead of exposing transport errors.
// Full presence diagnostics are implemented by later JSON tasks; this scaffold preserves the stable helper signature.
func DecodeLogicalTypeJSON(data []byte) (ColumnLogicalType, []SchemaValidationIssue) {
	var logical ColumnLogicalType
	if err := json.Unmarshal(data, &logical); err != nil {
		return logical, []SchemaValidationIssue{schemaValidationIssue("payload", SchemaIssueCodeValidationFailed, "logical type JSON payload must be a valid object")}
	}
	return logical, ValidateLogicalType(logical)
}
