package schema

// RelationMultiplicity represents the in-memory minimum and maximum child-row range for a table relation.
type RelationMultiplicity struct {
	// Min stores the minimum number of child rows generated for each upstream row.
	Min int `json:"min"`

	// Max stores the maximum number of child rows generated for each upstream row.
	Max int `json:"max"`
}

// NewRelationMultiplicity creates a relation multiplicity value and returns scaffold validation issues for the supplied range.
func NewRelationMultiplicity(min int, max int) (RelationMultiplicity, []SchemaValidationIssue) {
	return RelationMultiplicity{Min: min, Max: max}, ValidateRelationMultiplicity(min, max)
}

// ValidateRelationMultiplicity exposes the relation-domain validation entry point for multiplier ranges.
func ValidateRelationMultiplicity(min int, max int) []SchemaValidationIssue {
	return []SchemaValidationIssue{}
}
