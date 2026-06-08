package dialect

import "github.com/gerdong/loomidbx/internal/dbx/schema"

// Dialect renders SQL primitives without executing statements.
type Dialect interface {
	// QuoteIdentifier returns a safely quoted identifier according to database rules.
	QuoteIdentifier(name string) string

	// Placeholder returns the parameter placeholder for a one-based argument index.
	Placeholder(index int) string

	// BuildInsert returns SQL text and args for a batch insert request without executing it.
	BuildInsert(req InsertRequest) ([]Statement, error)
}

// InsertRequest describes a planned insert operation for SQL rendering.
type InsertRequest struct {
	// Schema stores the optional target schema or namespace.
	Schema string

	// Table stores the target table name.
	Table string

	// Columns stores canonical columns that define insert order.
	Columns []schema.Column

	// Rows stores row values keyed by column name.
	Rows []map[string]any
}

// Statement describes a parameterized SQL statement that a later writer may execute.
type Statement struct {
	// SQL stores the SQL text.
	SQL string

	// Args stores parameter values in placeholder order.
	Args []any
}
