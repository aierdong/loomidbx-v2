package schema

// RawMetadata stores database-specific metadata that is not normalized yet.
type RawMetadata map[string]any

// Database is the canonical root returned by introspection.
type Database struct {
	// Name stores the database name visible to the adapter.
	Name string

	// Catalog stores the catalog name when the database exposes one.
	Catalog string

	// Namespaces stores schema or namespace snapshots.
	Namespaces []Namespace

	// Raw stores unnormalized database metadata.
	Raw RawMetadata
}

// Namespace groups tables and views under a schema-like scope.
type Namespace struct {
	// Name stores the namespace or schema name.
	Name string

	// Tables stores table snapshots in deterministic order when provided by the introspector.
	Tables []Table

	// Views stores view snapshots in deterministic order when provided by the introspector.
	Views []View

	// Raw stores unnormalized namespace metadata.
	Raw RawMetadata
}

// Table describes a canonical table snapshot.
type Table struct {
	// Name stores the table name.
	Name string

	// Columns stores the table column snapshots.
	Columns []Column

	// PrimaryKey stores the primary key when present.
	PrimaryKey *PrimaryKey

	// ForeignKeys stores foreign key constraints.
	ForeignKeys []ForeignKey

	// UniqueConstraints stores unique constraints.
	UniqueConstraints []UniqueConstraint

	// CheckConstraints stores check constraints.
	CheckConstraints []CheckConstraint

	// Indexes stores index metadata.
	Indexes []Index

	// Comment stores an optional table comment.
	Comment string

	// Raw stores unnormalized table metadata.
	Raw RawMetadata
}

// View describes a canonical view snapshot.
type View struct {
	// Name stores the view name.
	Name string

	// Columns stores the view column snapshots.
	Columns []Column

	// Definition stores the view definition when safely available.
	Definition string

	// Comment stores an optional view comment.
	Comment string

	// Raw stores unnormalized view metadata.
	Raw RawMetadata
}

// Column describes a canonical column snapshot.
type Column struct {
	// Name stores the column name.
	Name string

	// NativeType stores the raw database type name or definition.
	NativeType string

	// LogicalType stores the normalized logical type.
	LogicalType LogicalType

	// Default stores a database default expression when available.
	Default string

	// Nullable reports whether null values are allowed.
	Nullable bool

	// Ordinal stores the one-based column position when known.
	Ordinal int

	// Generated reports whether the column is generated from an expression.
	Generated bool

	// Identity reports whether the column is database-generated identity data.
	Identity bool

	// AutoIncrement reports whether the column uses auto-increment semantics.
	AutoIncrement bool

	// Primary reports whether the column participates in the primary key.
	Primary bool

	// Unique reports whether the column has a single-column unique constraint.
	Unique bool

	// Comment stores an optional column comment.
	Comment string

	// Raw stores unnormalized column metadata.
	Raw RawMetadata
}
