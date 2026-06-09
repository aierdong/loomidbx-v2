package schema

// SchemaIdentity identifies a schema hierarchy for downstream consumers without exposing storage primary keys.
type SchemaIdentity struct {
	// ConnectionID stores the stable parent connection reference.
	ConnectionID int64 `json:"connectionId"`

	// CatalogName stores the database or catalog name portion of the identity.
	CatalogName string `json:"catalogName"`

	// SchemaName stores the schema name portion of the identity; an empty string represents an implicit schema.
	SchemaName string `json:"schemaName"`
}
