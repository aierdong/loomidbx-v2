package core

import (
	"context"

	"github.com/gerdong/loomidbx/internal/dbx/capability"
	"github.com/gerdong/loomidbx/internal/dbx/dialect"
	"github.com/gerdong/loomidbx/internal/dbx/introspect"
	"github.com/gerdong/loomidbx/internal/dbx/typex"
)

// DBType is the stable target database identifier.
type DBType string

const (
	// DBTypeMySQL identifies MySQL-compatible targets.
	DBTypeMySQL DBType = "mysql"

	// DBTypePostgres identifies PostgreSQL-compatible targets.
	DBTypePostgres DBType = "postgres"
)

// AdapterInfo exposes non-sensitive adapter metadata.
type AdapterInfo struct {
	// Type stores the adapter database type.
	Type DBType

	// DisplayName stores the user-facing adapter name.
	DisplayName string
}

// ConnectionConfig is a per-call boundary for target database connection data.
type ConnectionConfig struct {
	// Type stores the intended adapter database type.
	Type DBType

	// Host stores the target host when DSN is not the only input.
	Host string

	// Port stores the target port when applicable.
	Port int

	// Database stores the target database name.
	Database string

	// Schema stores the target schema or namespace.
	Schema string

	// Username stores the login name for this call.
	Username string

	// Password stores the secret for this call only and must not be logged or persisted by DBX contracts.
	Password string

	// DSN stores an optional driver-specific connection string for this call only.
	DSN string

	// Options stores non-standard connection options supplied at the boundary.
	Options map[string]string
}

// ConnectionTestResult reports connection validation without leaking driver details.
type ConnectionTestResult struct {
	// OK reports whether the adapter considered the connection valid.
	OK bool

	// Message stores a non-sensitive diagnostic message.
	Message string

	// Err stores an actionable failure cause when available.
	Err error
}

// Adapter is the unified database capability entry point.
type Adapter interface {
	// Type returns the stable database type identifier.
	Type() DBType

	// DisplayName returns non-sensitive adapter metadata for UI or diagnostics.
	DisplayName() string

	// Capabilities returns the database capability model used for runtime negotiation.
	Capabilities() capability.Capabilities

	// TestConnection validates a per-call connection config and returns a sanitized result.
	TestConnection(ctx context.Context, cfg ConnectionConfig) ConnectionTestResult

	// Dialect returns SQL rendering primitives without executing SQL.
	Dialect() dialect.Dialect

	// Introspector returns the metadata scanning boundary.
	Introspector() introspect.Introspector

	// TypeMapper returns the native-to-logical type mapping boundary.
	TypeMapper() typex.Mapper
}
