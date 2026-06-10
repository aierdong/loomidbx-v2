package schema

import "time"

// TableConstraintType identifies the stable table-level constraint category supported by this domain boundary.
type TableConstraintType string

const (
	// TableConstraintTypePrimary represents a PRIMARY key constraint over one or more ordered columns.
	TableConstraintTypePrimary TableConstraintType = "PRIMARY"

	// TableConstraintTypeUnique represents a UNIQUE constraint over one or more ordered columns.
	TableConstraintTypeUnique TableConstraintType = "UNIQUE"
)

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
