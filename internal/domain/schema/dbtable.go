package schema

import "time"

// DbTable represents a stable table domain entity under a persisted database schema.
type DbTable struct {
	// ID stores the persisted table primary key; draft objects may use zero before storage assigns an identity.
	ID int64 `json:"id"`

	// SchemaID stores the stable parent schema reference for this table.
	SchemaID int64 `json:"schemaId"`

	// TableName stores the stable database table name used by downstream schema consumers.
	TableName string `json:"tableName"`

	// Comment stores optional non-sensitive table documentation text.
	Comment string `json:"comment"`

	// DDLSnapshot stores an optional display snapshot of the table DDL without driving validation logic.
	DDLSnapshot string `json:"ddlSnapshot"`

	// ScannedAt stores the optional time when this table snapshot was scanned; nil means it has not been scanned.
	ScannedAt *time.Time `json:"scannedAt"`

	// CreatedAt stores the creation audit time for persisted table snapshots.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt stores the latest update audit time for persisted table snapshots.
	UpdatedAt time.Time `json:"updatedAt"`
}
