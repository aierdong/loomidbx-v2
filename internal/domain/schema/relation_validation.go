package schema

// ValidateForeignKey exposes the relation-domain validation entry point for ForeignKey models.
func ValidateForeignKey(foreignKey ForeignKey, mode SchemaValidationMode) []SchemaValidationIssue {
	return []SchemaValidationIssue{}
}

// ValidateTableRelation exposes the relation-domain validation entry point for TableRelation models.
func ValidateTableRelation(relation TableRelation, mode SchemaValidationMode) []SchemaValidationIssue {
	return []SchemaValidationIssue{}
}

// ValidateRelationType exposes the relation-domain validation entry point for RelationType values.
func ValidateRelationType(relationType RelationType) []SchemaValidationIssue {
	return []SchemaValidationIssue{}
}
