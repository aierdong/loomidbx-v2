package schema

import "fmt"

// ValidateForeignKey exposes the relation-domain validation entry point for ForeignKey models.
func ValidateForeignKey(foreignKey ForeignKey, mode SchemaValidationMode) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 8)
	if mode.IsUnknown() {
		issues = append(issues, schemaValidationIssue("mode", SchemaIssueCodeValidationFailed, "validation mode must be draft or persisted"))
	}

	if !mode.IsUnknown() {
		issues = append(issues, validateSnapshotID("foreign key", foreignKey.ID, mode)...)
	}
	if foreignKey.TableID <= 0 {
		issues = append(issues, schemaValidationIssue("tableId", SchemaIssueCodeInvalidID, "tableId must reference a saved table"))
	}
	issues = append(issues, validateRequiredName("fkName", foreignKey.FKName)...)
	if foreignKey.ReferencedTableID <= 0 {
		issues = append(issues, schemaValidationIssue("referencedTableId", SchemaIssueCodeInvalidID, "referencedTableId must reference a saved table"))
	}
	issues = append(issues, validateRelationColumnMapping("columnIds", foreignKey.ColumnIDs, "referencedColumnIds", foreignKey.ReferencedColumnIDs)...)
	if !mode.IsUnknown() && mode == SchemaValidationModePersisted && foreignKey.CreatedAt.IsZero() {
		issues = append(issues, schemaValidationIssue("createdAt", SchemaIssueCodeInvalidTime, "createdAt is required for persisted foreign key snapshots"))
	}

	return issues
}

// ValidateTableRelation exposes the relation-domain validation entry point for TableRelation models.
func ValidateTableRelation(relation TableRelation, mode SchemaValidationMode) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 10)
	if mode.IsUnknown() {
		issues = append(issues, schemaValidationIssue("mode", SchemaIssueCodeValidationFailed, "validation mode must be draft or persisted"))
	}

	if !mode.IsUnknown() {
		issues = append(issues, validateSnapshotID("table relation", relation.ID, mode)...)
	}
	issues = append(issues, ValidateRelationType(relation.RelationType)...)
	if relation.ParentTableID <= 0 {
		issues = append(issues, schemaValidationIssue("parentTableId", SchemaIssueCodeInvalidID, "parentTableId must reference a saved table"))
	}
	if relation.ChildTableID <= 0 {
		issues = append(issues, schemaValidationIssue("childTableId", SchemaIssueCodeInvalidID, "childTableId must reference a saved table"))
	}
	issues = append(issues, validateRelationColumnMapping("parentColumnIds", relation.ParentColumnIDs, "childColumnIds", relation.ChildColumnIDs)...)
	issues = append(issues, ValidateRelationMultiplicity(relation.MultiplierMin, relation.MultiplierMax)...)
	issues = append(issues, validateAuditTimes(mode, relation.CreatedAt, relation.UpdatedAt)...)

	return issues
}

// ValidateRelationType exposes the relation-domain validation entry point for RelationType values.
func ValidateRelationType(relationType RelationType) []SchemaValidationIssue {
	if relationType.IsKnown() {
		return nil
	}
	return []SchemaValidationIssue{schemaValidationIssue("relationType", SchemaIssueCodeInvalidType, "relationType must be PARENT_CHILD or JOIN_TABLE")}
}

func validateRelationColumnMapping(leftPath string, leftIDs []int64, rightPath string, rightIDs []int64) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 4)
	if len(leftIDs) == 0 {
		issues = append(issues, schemaValidationIssue(leftPath, SchemaIssueCodeRequired, leftPath+" is required and must contain at least one column id"))
	}
	if len(rightIDs) == 0 {
		issues = append(issues, schemaValidationIssue(rightPath, SchemaIssueCodeRequired, rightPath+" is required and must contain at least one column id"))
	}

	issues = append(issues, validateRelationColumnIDs(leftPath, leftIDs)...)
	issues = append(issues, validateRelationColumnIDs(rightPath, rightIDs)...)
	if len(leftIDs) != len(rightIDs) {
		issues = append(issues, schemaValidationIssue(rightPath, SchemaIssueCodeInvalidMapping, rightPath+" must contain the same number of column ids as "+leftPath))
	}
	return issues
}

func validateRelationColumnIDs(path string, columnIDs []int64) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 2)
	seen := make(map[int64]struct{}, len(columnIDs))
	duplicateID := false
	for index, columnID := range columnIDs {
		if columnID <= 0 {
			issues = append(issues, schemaValidationIssue(fmt.Sprintf("%s[%d]", path, index), SchemaIssueCodeInvalidID, path+" must contain only saved column references"))
			continue
		}
		if _, ok := seen[columnID]; ok {
			duplicateID = true
			continue
		}
		seen[columnID] = struct{}{}
	}
	if duplicateID {
		issues = append(issues, schemaValidationIssue(path, SchemaIssueCodeInvalidMapping, path+" must not contain duplicate column references"))
	}
	return issues
}
