package schema

// ValidateTable validates a table model using the explicit draft or persisted schema validation mode.
// Full table validation rules are implemented by later validation tasks; this scaffold only preserves the stable entry point.
func ValidateTable(table DbTable, mode SchemaValidationMode) []SchemaValidationIssue {
	if mode.IsUnknown() {
		return []SchemaValidationIssue{schemaValidationIssue("mode", SchemaIssueCodeValidationFailed, "validation mode must be draft or persisted")}
	}
	return nil
}

// ValidateColumn validates a column model using the explicit draft or persisted schema validation mode.
// Full column validation rules are implemented by later validation tasks; this scaffold only preserves the stable entry point.
func ValidateColumn(column DbColumn, mode SchemaValidationMode) []SchemaValidationIssue {
	if mode.IsUnknown() {
		return []SchemaValidationIssue{schemaValidationIssue("mode", SchemaIssueCodeValidationFailed, "validation mode must be draft or persisted")}
	}
	return nil
}

// ValidateConstraint validates a table constraint model using the explicit draft or persisted schema validation mode.
// Full constraint validation rules are implemented by later validation tasks; this scaffold only preserves the stable entry point.
func ValidateConstraint(constraint TableConstraint, mode SchemaValidationMode) []SchemaValidationIssue {
	if mode.IsUnknown() {
		return []SchemaValidationIssue{schemaValidationIssue("mode", SchemaIssueCodeValidationFailed, "validation mode must be draft or persisted")}
	}
	return nil
}

// ValidateLogicalType validates logical type metadata without applying draft or persisted snapshot rules.
// Full logical type validation rules are implemented by later validation tasks; this scaffold only preserves the stable entry point.
func ValidateLogicalType(logical ColumnLogicalType) []SchemaValidationIssue {
	return nil
}
