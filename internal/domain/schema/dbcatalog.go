package schema

import "time"

// DbCatalog represents a scanned database or catalog level under a saved connection.
type DbCatalog struct {
	// ID stores the persisted catalog primary key; draft objects may use zero before storage assigns an identity.
	ID int64 `json:"id"`

	// ConnectionID stores the stable parent connection reference owned by the connection domain.
	ConnectionID int64 `json:"connectionId"`

	// CatalogName stores the database or catalog name used as the stable catalog-level name.
	CatalogName string `json:"catalogName"`

	// ScannedAt stores the optional time when this catalog snapshot was scanned; nil means it has not been scanned.
	ScannedAt *time.Time `json:"scannedAt"`

	// CreatedAt stores the creation audit time for persisted catalog snapshots.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt stores the latest update audit time for persisted catalog snapshots.
	UpdatedAt time.Time `json:"updatedAt"`
}
