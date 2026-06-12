package execution

import "time"

// ExecutionTableResult is the table-level execution history record with table and schema name snapshots.
type ExecutionTableResult struct {
	// ID stores the stable table result identity; new unsaved results may use zero and loaded history uses a positive value.
	ID int64 `json:"id"`

	// ExecutionTaskID stores the parent execution task identity for this table result.
	ExecutionTaskID int64 `json:"executionTaskId"`

	// TableID stores the optional referenced schema table identity; nil preserves history when the source table was deleted.
	TableID *int64 `json:"tableId,omitempty"`

	// TableNameSnapshot stores the table name captured at execution time.
	TableNameSnapshot string `json:"tableNameSnapshot"`

	// SchemaNameSnapshot stores the schema name captured at execution time.
	SchemaNameSnapshot string `json:"schemaNameSnapshot"`

	// RowsWritten stores the number of rows successfully written for this table.
	RowsWritten int64 `json:"rowsWritten"`

	// Status stores the current or final table-level status as a stable string-backed enum.
	Status ExecutionTableStatus `json:"status"`

	// ErrorSnapshot stores an optional safe error summary for failed table execution.
	ErrorSnapshot *ExecutionErrorSnapshot `json:"errorSnapshot,omitempty"`

	// ExecutionOrder stores the actual one-based execution order for this table result.
	ExecutionOrder int `json:"executionOrder"`

	// CreatedAt stores when this table result record was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt stores when this table result record was last updated.
	UpdatedAt time.Time `json:"updatedAt"`
}
