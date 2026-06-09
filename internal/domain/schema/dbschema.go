package schema

import "time"

// DbSchema represents a schema level under a catalog, including the empty-string implicit schema form.
type DbSchema struct {
	// ID stores the persisted schema primary key; draft objects may use zero before storage assigns an identity.
	ID int64 `json:"id"`

	// CatalogID stores the stable parent catalog reference for this schema.
	CatalogID int64 `json:"catalogId"`

	// SchemaName stores the schema name; an empty string is the stable implicit schema representation.
	SchemaName string `json:"schemaName"`

	// ScannedAt stores the optional time when this schema snapshot was scanned; nil means it has not been scanned.
	ScannedAt *time.Time `json:"scannedAt"`

	// CreatedAt stores the creation audit time for persisted schema snapshots.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt stores the latest update audit time for persisted schema snapshots.
	UpdatedAt time.Time `json:"updatedAt"`
}
