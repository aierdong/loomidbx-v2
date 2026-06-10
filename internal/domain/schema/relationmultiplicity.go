package schema

// RelationMultiplicity expresses the allowed child-row count range per upstream relation row.
type RelationMultiplicity struct {
	// Min stores the minimum generated child rows per upstream row and must be non-negative.
	Min int `json:"min"`

	// Max stores the maximum generated child rows per upstream row and must be greater than or equal to Min.
	Max int `json:"max"`
}

// NewRelationMultiplicity creates a relation multiplicity value object and returns field-level issues for invalid ranges.
func NewRelationMultiplicity(min int, max int) (RelationMultiplicity, []SchemaValidationIssue) {
	multiplicity := RelationMultiplicity{Min: min, Max: max}
	return multiplicity, ValidateRelationMultiplicity(min, max)
}

// ValidateRelationMultiplicity validates the multiplierMin and multiplierMax range contract for table relations.
func ValidateRelationMultiplicity(min int, max int) []SchemaValidationIssue {
	issues := make([]SchemaValidationIssue, 0, 2)
	if min < 0 {
		issues = append(issues, schemaValidationIssue("multiplierMin", SchemaIssueCodeInvalidRange, "multiplierMin must be greater than or equal to zero"))
	}
	if max < min {
		issues = append(issues, schemaValidationIssue("multiplierMax", SchemaIssueCodeInvalidRange, "multiplierMax must be greater than or equal to multiplierMin"))
	}
	return issues
}
