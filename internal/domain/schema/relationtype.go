package schema

import "encoding/json"

// RelationType identifies the stable domain category for a table relation.
type RelationType string

const (
	// RelationTypeParentChild represents a classic one-way parent-to-child dependency relation.
	RelationTypeParentChild RelationType = "PARENT_CHILD"

	// RelationTypeJoinTable represents one flattened base-table-to-join-table branch of an N:N relation.
	RelationTypeJoinTable RelationType = "JOIN_TABLE"
)

// IsKnown reports whether the relation type belongs to the stable relation type set.
func (t RelationType) IsKnown() bool {
	switch t {
	case RelationTypeParentChild, RelationTypeJoinTable:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the relation type is outside the stable relation type set.
func (t RelationType) IsUnknown() bool {
	return !t.IsKnown()
}

// String returns the stable string representation used for persistence and transport.
func (t RelationType) String() string {
	return string(t)
}

// MarshalJSON serializes the relation type as its stable string value.
func (t RelationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON restores the relation type from its serialized string value.
func (t *RelationType) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*t = RelationType(value)
	return nil
}
