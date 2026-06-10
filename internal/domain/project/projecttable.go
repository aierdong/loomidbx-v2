package project

import "time"

// ProjectTable is a table-level execution configuration snapshot within a persisted Project.
type ProjectTable struct {
	// ID stores the stable ProjectTable identity; drafts use zero and persisted ProjectTables use a positive value.
	ID int64 `json:"id"`

	// ProjectID stores the owning persisted Project identity and must be positive.
	ProjectID int64 `json:"projectId"`

	// TableID stores the target schema table identity and must be positive.
	TableID int64 `json:"tableId"`

	// RowCount stores the nullable row target; nil, zero, and positive values have distinct meanings.
	RowCount *int `json:"rowCount"`

	// TruncateBefore stores whether the target table should be cleared before execution.
	TruncateBefore bool `json:"truncateBefore"`

	// ExecutionOrder stores the precomputed execution order snapshot for this table.
	ExecutionOrder int `json:"executionOrder"`

	// CreatedAt stores when the persisted ProjectTable was created; drafts may keep the zero time.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt stores when the persisted ProjectTable was last changed; drafts may keep the zero time.
	UpdatedAt time.Time `json:"updatedAt"`
}
