package schema

import (
	"encoding/json"
	"time"
)

// TableConstraintType identifies the stable table-level constraint category supported by this domain boundary.
type TableConstraintType string

const (
	// TableConstraintTypePrimary represents a PRIMARY key constraint over one or more ordered columns.
	TableConstraintTypePrimary TableConstraintType = "PRIMARY"

	// TableConstraintTypeUnique represents a UNIQUE constraint over one or more ordered columns.
	TableConstraintTypeUnique TableConstraintType = "UNIQUE"
)

// IsKnown reports whether the table constraint type belongs to the stable supported set.
func (constraintType TableConstraintType) IsKnown() bool {
	switch constraintType {
	case TableConstraintTypePrimary,
		TableConstraintTypeUnique:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the table constraint type is outside the stable supported set.
func (constraintType TableConstraintType) IsUnknown() bool {
	return !constraintType.IsKnown()
}

// String returns the stable string representation used for persistence and transport.
func (constraintType TableConstraintType) String() string {
	return string(constraintType)
}

// MarshalJSON serializes the table constraint type as its stable string value.
func (constraintType TableConstraintType) MarshalJSON() ([]byte, error) {
	return json.Marshal(constraintType.String())
}

// UnmarshalJSON restores the table constraint type from its serialized string value.
func (constraintType *TableConstraintType) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*constraintType = TableConstraintType(value)
	return nil
}

// TableConstraint represents a PRIMARY or UNIQUE table-level constraint with ordered column references.
type TableConstraint struct {
	// ID stores the persisted constraint primary key; draft objects may use zero before storage assigns an identity.
	ID int64 `json:"id"`

	// TableID stores the stable parent table reference for this constraint.
	TableID int64 `json:"tableId"`

	// ConstraintName stores the stable constraint name generated or discovered by later mapping layers.
	ConstraintName string `json:"constraintName"`

	// ConstraintType stores whether the constraint is PRIMARY or UNIQUE.
	ConstraintType TableConstraintType `json:"constraintType"`

	// ColumnIDs stores ordered column identifiers; order represents the database constraint column order.
	ColumnIDs []int64 `json:"columnIds"`

	// CreatedAt stores the creation audit time for persisted constraint snapshots.
	CreatedAt time.Time `json:"createdAt"`
}
