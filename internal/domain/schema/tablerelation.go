package schema

import "time"

// TableRelation represents a single explicit table relationship used by downstream generation planning.
type TableRelation struct {
	// ID stores the persisted table relation primary key; draft objects may use zero before storage assigns an identity.
	ID int64 `json:"id"`

	// RelationType stores whether this relation is a parent-child dependency or a join-table branch.
	RelationType RelationType `json:"relationType"`

	// ParentTableID stores the parent table reference, or the base table reference for join-table relations.
	ParentTableID int64 `json:"parentTableId"`

	// ChildTableID stores the child table reference, or the join table reference for join-table relations.
	ChildTableID int64 `json:"childTableId"`

	// ParentColumnIDs stores ordered parent or base-table column identifiers for the relation mapping.
	ParentColumnIDs []int64 `json:"parentColumnIds"`

	// ChildColumnIDs stores ordered child or join-table column identifiers paired with ParentColumnIDs by index.
	ChildColumnIDs []int64 `json:"childColumnIds"`

	// MultiplierMin stores the minimum number of child rows generated for each upstream row.
	MultiplierMin int `json:"multiplierMin"`

	// MultiplierMax stores the maximum number of child rows generated for each upstream row.
	MultiplierMax int `json:"multiplierMax"`

	// IsLogical reports whether the relation is application-defined instead of derived directly from a physical foreign key.
	IsLogical bool `json:"isLogical"`

	// CreatedAt stores the creation audit time for persisted table relation snapshots.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt stores the latest update audit time for persisted table relation snapshots.
	UpdatedAt time.Time `json:"updatedAt"`
}
