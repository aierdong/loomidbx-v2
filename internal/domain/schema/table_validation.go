package schema

import "strings"

// ValidateTable validates a table model using the explicit draft or persisted schema validation mode.
func ValidateTable(table DbTable, mode SchemaValidationMode) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 6)
	if mode.IsUnknown() {
		issues = append(issues, schemaValidationIssue("mode", SchemaIssueCodeValidationFailed, "validation mode must be draft or persisted"))
	}

	if !mode.IsUnknown() {
		issues = append(issues, validateSnapshotID("table", table.ID, mode)...)
	}
	if table.SchemaID <= 0 {
		issues = append(issues, schemaValidationIssue("schemaId", SchemaIssueCodeInvalidID, "schemaId must reference a saved schema"))
	}
	issues = append(issues, validateRequiredName("tableName", table.TableName)...)
	issues = append(issues, validateOptionalTime("scannedAt", table.ScannedAt)...)
	issues = append(issues, validateAuditTimes(mode, table.CreatedAt, table.UpdatedAt)...)

	return issues
}

// ValidateColumn validates a column model using the explicit draft or persisted schema validation mode.
func ValidateColumn(column DbColumn, mode SchemaValidationMode) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 8)
	if mode.IsUnknown() {
		issues = append(issues, schemaValidationIssue("mode", SchemaIssueCodeValidationFailed, "validation mode must be draft or persisted"))
	}

	if !mode.IsUnknown() {
		issues = append(issues, validateSnapshotID("column", column.ID, mode)...)
	}
	if column.TableID <= 0 {
		issues = append(issues, schemaValidationIssue("tableId", SchemaIssueCodeInvalidID, "tableId must reference a saved table"))
	}
	if column.OrdinalPosition <= 0 {
		issues = append(issues, schemaValidationIssue("ordinalPosition", SchemaIssueCodeValidationFailed, "ordinalPosition must be greater than zero"))
	}
	issues = append(issues, validateRequiredName("columnName", column.ColumnName)...)
	if strings.TrimSpace(column.NativeType) == "" {
		issues = append(issues, schemaValidationIssue("nativeType", SchemaIssueCodeRequired, "nativeType is required"))
	}
	issues = append(issues, prefixedIssues("logicalType", ValidateLogicalType(column.LogicalType))...)
	issues = append(issues, validateAuditTimes(mode, column.CreatedAt, column.UpdatedAt)...)

	return issues
}

// ValidateConstraint validates a table constraint model using the explicit draft or persisted schema validation mode.
func ValidateConstraint(constraint TableConstraint, mode SchemaValidationMode) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 7)
	if mode.IsUnknown() {
		issues = append(issues, schemaValidationIssue("mode", SchemaIssueCodeValidationFailed, "validation mode must be draft or persisted"))
	}

	if !mode.IsUnknown() {
		issues = append(issues, validateSnapshotID("constraint", constraint.ID, mode)...)
	}
	if constraint.TableID <= 0 {
		issues = append(issues, schemaValidationIssue("tableId", SchemaIssueCodeInvalidID, "tableId must reference a saved table"))
	}
	issues = append(issues, validateRequiredName("constraintName", constraint.ConstraintName)...)
	if constraint.ConstraintType.IsUnknown() {
		issues = append(issues, schemaValidationIssue("constraintType", SchemaIssueCodeValidationFailed, "constraintType must be PRIMARY or UNIQUE"))
	}
	issues = append(issues, validateConstraintColumnIDs(constraint.ColumnIDs)...)
	if !mode.IsUnknown() && mode == SchemaValidationModePersisted && constraint.CreatedAt.IsZero() {
		issues = append(issues, schemaValidationIssue("createdAt", SchemaIssueCodeInvalidTime, "createdAt is required for persisted constraint snapshots"))
	}

	return issues
}

// ValidateLogicalType validates logical type metadata without applying draft or persisted snapshot rules.
func ValidateLogicalType(logical ColumnLogicalType) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 6)
	if logical.Kind.IsUnknown() {
		issues = append(issues, schemaValidationIssue("kind", SchemaIssueCodeValidationFailed, "kind must be a known logical type"))
	}
	if logical.Length != nil && *logical.Length <= 0 {
		issues = append(issues, schemaValidationIssue("length", SchemaIssueCodeValidationFailed, "length must be greater than zero"))
	}
	if logical.Precision != nil && *logical.Precision <= 0 {
		issues = append(issues, schemaValidationIssue("precision", SchemaIssueCodeValidationFailed, "precision must be greater than zero"))
	}
	if logical.Scale != nil && *logical.Scale < 0 {
		issues = append(issues, schemaValidationIssue("scale", SchemaIssueCodeValidationFailed, "scale must not be negative"))
	} else if logical.Scale != nil && logical.Precision != nil && *logical.Scale > *logical.Precision && *logical.Precision > 0 {
		issues = append(issues, schemaValidationIssue("scale", SchemaIssueCodeValidationFailed, "scale must not be greater than precision"))
	}
	if logical.BitWidth != nil && *logical.BitWidth <= 0 {
		issues = append(issues, schemaValidationIssue("bitWidth", SchemaIssueCodeValidationFailed, "bitWidth must be greater than zero"))
	}
	if logical.Kind == ColumnLogicalKindUnknown && strings.TrimSpace(logical.NativeType) == "" {
		issues = append(issues, schemaValidationIssue("nativeType", SchemaIssueCodeRequired, "nativeType is required for unknown logical types"))
	}
	if logical.Kind == ColumnLogicalKindArray && logical.Element == nil {
		issues = append(issues, schemaValidationIssue("element", SchemaIssueCodeRequired, "element is required for array logical types"))
	}
	if logical.Kind == ColumnLogicalKindEnum {
		issues = append(issues, validateEnumValues(logical.EnumValues)...)
	}

	return issues
}

func validateSnapshotID(kind string, id int64, mode SchemaValidationMode) []SchemaValidationIssue {
	if mode == SchemaValidationModePersisted {
		if id <= 0 {
			return []SchemaValidationIssue{schemaValidationIssue("id", SchemaIssueCodeInvalidID, "id must be greater than zero for persisted "+kind+" snapshots")}
		}
	} else if id < 0 {
		return []SchemaValidationIssue{schemaValidationIssue("id", SchemaIssueCodeInvalidID, "id must not be negative for draft "+kind+" objects")}
	}
	return nil
}

func validateConstraintColumnIDs(columnIDs []int64) []SchemaValidationIssue {
	if len(columnIDs) == 0 {
		return []SchemaValidationIssue{schemaValidationIssue("columnIds", SchemaIssueCodeRequired, "columnIds is required and must contain at least one column id")}
	}

	issues := make([]SchemaValidationIssue, 0, 2)
	seen := make(map[int64]struct{}, len(columnIDs))
	invalidID := false
	duplicateID := false
	for _, columnID := range columnIDs {
		if columnID <= 0 {
			invalidID = true
			continue
		}
		if _, ok := seen[columnID]; ok {
			duplicateID = true
			continue
		}
		seen[columnID] = struct{}{}
	}
	if invalidID {
		issues = append(issues, schemaValidationIssue("columnIds", SchemaIssueCodeInvalidID, "columnIds must contain only saved column references"))
	}
	if duplicateID {
		issues = append(issues, schemaValidationIssue("columnIds", SchemaIssueCodeValidationFailed, "columnIds must not contain duplicate column references"))
	}
	return issues
}

func validateEnumValues(values []string) []SchemaValidationIssue {
	if len(values) == 0 {
		return []SchemaValidationIssue{schemaValidationIssue("enumValues", SchemaIssueCodeRequired, "enumValues is required for enum logical types")}
	}

	issues := make([]SchemaValidationIssue, 0, 2)
	seen := make(map[string]struct{}, len(values))
	blankValue := false
	duplicateValue := false
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			blankValue = true
			continue
		}
		if _, ok := seen[trimmed]; ok {
			duplicateValue = true
			continue
		}
		seen[trimmed] = struct{}{}
	}
	if blankValue {
		issues = append(issues, schemaValidationIssue("enumValues", SchemaIssueCodeRequired, "enumValues must not contain blank values"))
	}
	if duplicateValue {
		issues = append(issues, schemaValidationIssue("enumValues", SchemaIssueCodeValidationFailed, "enumValues must not contain duplicate values"))
	}
	return issues
}

func prefixedIssues(prefix string, issues []SchemaValidationIssue) []SchemaValidationIssue {
	if len(issues) == 0 {
		return nil
	}
	prefixed := make([]SchemaValidationIssue, len(issues))
	for index, issue := range issues {
		issue.Path = prefix + "." + issue.Path
		prefixed[index] = issue
	}
	return prefixed
}
