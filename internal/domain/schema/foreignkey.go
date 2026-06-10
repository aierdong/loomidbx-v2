package schema

import "time"

// ForeignKey represents a physical foreign key constraint discovered for a database table.
type ForeignKey struct {
	// ID stores the persisted foreign key primary key; draft objects may use zero before storage assigns an identity.
	ID int64 `json:"id"`

	// TableID stores the source table reference where the foreign key columns are defined.
	TableID int64 `json:"tableId"`

	// FKName stores the stable database foreign key name used by downstream schema consumers.
	FKName string `json:"fkName"`

	// ReferencedTableID stores the target table reference that the foreign key points to.
	ReferencedTableID int64 `json:"referencedTableId"`

	// ColumnIDs stores ordered source column identifiers for the foreign key column mapping.
	ColumnIDs []int64 `json:"columnIds"`

	// ReferencedColumnIDs stores ordered target column identifiers paired with ColumnIDs by index.
	ReferencedColumnIDs []int64 `json:"referencedColumnIds"`

	// CreatedAt stores the creation audit time for persisted foreign key snapshots.
	CreatedAt time.Time `json:"createdAt"`
}
