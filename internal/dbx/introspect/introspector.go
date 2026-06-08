package introspect

import (
	"context"

	"github.com/gerdong/loomidbx/internal/dbx/schema"
)

// Connection is the narrow query boundary used by introspection.
type Connection interface {
	// QueryContext runs a metadata query through a caller-provided connection or fake.
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
}

// Rows is the narrow row iteration boundary used by introspection.
type Rows interface {
	// Close releases row resources.
	Close() error

	// Err returns the terminal row iteration error.
	Err() error

	// Next advances to the next row.
	Next() bool

	// Scan copies the current row into destination values.
	Scan(dest ...any) error
}

// Options scopes metadata scanning.
type Options struct {
	// Catalog limits scanning to a database catalog when applicable.
	Catalog string

	// Schema limits scanning to a schema or namespace when applicable.
	Schema string

	// Tables limits scanning to specific table names when supplied.
	Tables []string
}

// Introspector scans metadata into canonical schema without owning connections.
type Introspector interface {
	// Introspect returns a canonical schema snapshot or a typed failure.
	Introspect(ctx context.Context, conn Connection, opts Options) (*schema.Database, error)
}
