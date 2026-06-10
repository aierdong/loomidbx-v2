package schema

import "encoding/json"

// RelationType identifies the stable relationship category for a TableRelation.
type RelationType string

const (
	// RelationTypeParentChild represents a classic parent-to-child dependency relation.
	RelationTypeParentChild RelationType = "PARENT_CHILD"

	// RelationTypeJoinTable represents one flattened base-table to join-table dependency branch.
	RelationTypeJoinTable RelationType = "JOIN_TABLE"
)

// IsKnown reports whether the relation type belongs to the stable supported set.
func (relationType RelationType) IsKnown() bool {
	switch relationType {
	case RelationTypeParentChild,
		RelationTypeJoinTable:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the relation type is outside the stable supported set.
func (relationType RelationType) IsUnknown() bool {
	return !relationType.IsKnown()
}

// String returns the stable string representation used for persistence and transport.
func (relationType RelationType) String() string {
	return string(relationType)
}

// MarshalJSON serializes the relation type as its stable string value.
func (relationType RelationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(relationType.String())
}

// UnmarshalJSON restores the relation type from its serialized string value.
func (relationType *RelationType) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*relationType = RelationType(value)
	return nil
}
