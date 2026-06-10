package rule

import "time"

// GeneratorConfig represents a field-level generator binding owned by the schema domain.
type GeneratorConfig struct {
	// ID stores the persisted generator config primary key; draft objects may use zero before storage assigns an identity.
	ID int64 `json:"id"`

	// ColumnID stores the stable parent DbColumn reference for this generator configuration.
	ColumnID int64 `json:"columnId"`

	// GeneratorName stores the stable generator identifier selected by a later generator registry boundary.
	GeneratorName string `json:"generatorName"`

	// DataMappingType stores the logical output mapping category declared for generated values.
	DataMappingType DataMappingType `json:"dataMappingType"`

	// Params stores the generator-specific JSON payload without binding the domain layer to a concrete generator schema.
	Params GeneratorParams `json:"params"`

	// ConfigStatus stores whether this generator configuration is currently usable or requires review.
	ConfigStatus ConfigStatus `json:"configStatus"`

	// CreatedAt stores the creation audit time for persisted generator config snapshots.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt stores the latest update audit time for persisted generator config snapshots.
	UpdatedAt time.Time `json:"updatedAt"`
}
