package schema

// ValidateTable validates a table model using the explicit draft or persisted schema validation mode.
// Full table validation rules are implemented by later validation tasks; this scaffold only preserves the stable entry point.
func ValidateTable(table DbTable, mode SchemaValidationMode) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 1)
	issues = append(issues, validateTableSchemaMode(mode)...)
	return issues
}

// ValidateColumn validates a column model using the explicit draft or persisted schema validation mode.
// Full column validation rules are implemented by later validation tasks; this scaffold only preserves the stable entry point.
func ValidateColumn(column DbColumn, mode SchemaValidationMode) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 1)
	issues = append(issues, validateTableSchemaMode(mode)...)
	return issues
}

// ValidateConstraint validates a table constraint model using the explicit draft or persisted schema validation mode.
// Full constraint validation rules are implemented by later validation tasks; this scaffold only preserves the stable entry point.
func ValidateConstraint(constraint TableConstraint, mode SchemaValidationMode) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 2)
	issues = append(issues, validateTableSchemaMode(mode)...)
	if constraint.ConstraintType.IsUnknown() {
		issues = append(issues, schemaValidationIssue("constraintType", SchemaIssueCodeValidationFailed, "constraintType must be PRIMARY or UNIQUE"))
	}
	return issues
}

// ValidateLogicalType validates logical type metadata without applying draft or persisted snapshot rules.
// Full logical type validation rules are implemented by later validation tasks; this scaffold only preserves the stable entry point.
func ValidateLogicalType(logical ColumnLogicalType) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 1)
	if logical.Kind.IsUnknown() {
		issues = append(issues, schemaValidationIssue("kind", SchemaIssueCodeValidationFailed, "kind must be a known logical type"))
	}
	return issues
}

func validateTableSchemaMode(mode SchemaValidationMode) []SchemaValidationIssue {
	if mode.IsUnknown() {
		return []SchemaValidationIssue{schemaValidationIssue("mode", SchemaIssueCodeValidationFailed, "validation mode must be draft or persisted")}
	}
	return nil
}
