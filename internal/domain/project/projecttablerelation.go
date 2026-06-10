package project

import "time"

// ProjectTableRelation is a relation instantiation snapshot within a persisted Project.
type ProjectTableRelation struct {
	// ID stores the stable ProjectTableRelation identity; drafts use zero and persisted relations use a positive value.
	ID int64 `json:"id"`

	// ProjectID stores the owning persisted Project identity and must be positive.
	ProjectID int64 `json:"projectId"`

	// TableRelationID stores the referenced Schema-layer TableRelation identity.
	TableRelationID int64 `json:"tableRelationId"`

	// ParentProjectTableID stores the optional upstream ProjectTable identity; nil means the upstream table is outside this Project.
	ParentProjectTableID *int64 `json:"parentProjectTableId"`

	// ChildProjectTableID stores the downstream ProjectTable identity and must be present for an instantiated relation.
	ChildProjectTableID int64 `json:"childProjectTableId"`

	// MultiplierMin stores the minimum child-row multiplier copied from the Schema relation snapshot.
	MultiplierMin int `json:"multiplierMin"`

	// MultiplierMax stores the maximum child-row multiplier copied from the Schema relation snapshot.
	MultiplierMax int `json:"multiplierMax"`

	// RelValueSource stores how relation values are read for this relation configuration snapshot.
	RelValueSource RelationValueSource `json:"relValueSource"`

	// RelSourceSQL stores optional SQL text used by database-query or merged value sources; the domain model never executes it.
	RelSourceSQL string `json:"relSourceSql"`

	// CreatedAt stores when the persisted ProjectTableRelation was created; drafts may keep the zero time.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt stores when the persisted ProjectTableRelation was last changed; drafts may keep the zero time.
	UpdatedAt time.Time `json:"updatedAt"`
}
