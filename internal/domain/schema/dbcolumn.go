package schema

import "time"

// DbColumn represents a stable column domain entity under a database table.
type DbColumn struct {
	// ID stores the persisted column primary key; draft objects may use zero before storage assigns an identity.
	ID int64 `json:"id"`

	// TableID stores the stable parent table reference for this column.
	TableID int64 `json:"tableId"`

	// OrdinalPosition stores the one-based column order within the parent table.
	OrdinalPosition int `json:"ordinalPosition"`

	// ColumnName stores the stable database column name used by downstream schema consumers.
	ColumnName string `json:"columnName"`

	// NativeType stores the database-native type text for diagnostics and future mapping.
	NativeType string `json:"nativeType"`

	// LogicalType stores the stable logical type metadata consumed by rules and generators.
	LogicalType ColumnLogicalType `json:"logicalType"`

	// Nullable reports whether the database column accepts NULL values.
	Nullable bool `json:"nullable"`

	// DefaultValue stores the raw database default expression; nil means the column has no default.
	DefaultValue *string `json:"defaultValue"`

	// IsPrimaryKey stores a derived primary-key flag for read optimization; synchronization is owned by later tasks.
	IsPrimaryKey bool `json:"isPrimaryKey"`

	// Comment stores optional non-sensitive column documentation text.
	Comment string `json:"comment"`

	// CreatedAt stores the creation audit time for persisted column snapshots.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt stores the latest update audit time for persisted column snapshots.
	UpdatedAt time.Time `json:"updatedAt"`
}
