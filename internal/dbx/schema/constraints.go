package schema

// PrimaryKey describes a table primary key constraint.
type PrimaryKey struct {
	// Name stores the constraint name when available.
	Name string

	// Columns stores constrained column names in key order.
	Columns []string

	// Raw stores unnormalized primary key metadata.
	Raw RawMetadata
}

// ForeignKey describes a foreign key relationship.
type ForeignKey struct {
	// Name stores the constraint name when available.
	Name string

	// Columns stores referencing column names.
	Columns []string

	// ReferencedNamespace stores the referenced namespace when available.
	ReferencedNamespace string

	// ReferencedTable stores the referenced table name.
	ReferencedTable string

	// ReferencedColumns stores referenced column names.
	ReferencedColumns []string

	// OnUpdate stores the update action when available.
	OnUpdate string

	// OnDelete stores the delete action when available.
	OnDelete string

	// Deferrable reports whether the foreign key can be deferred.
	Deferrable bool

	// InitiallyDeferred reports whether the foreign key starts deferred.
	InitiallyDeferred bool

	// Raw stores unnormalized foreign key metadata.
	Raw RawMetadata
}

// UniqueConstraint describes a unique constraint.
type UniqueConstraint struct {
	// Name stores the constraint name when available.
	Name string

	// Columns stores constrained column names.
	Columns []string

	// Raw stores unnormalized unique constraint metadata.
	Raw RawMetadata
}

// CheckConstraint describes a check constraint.
type CheckConstraint struct {
	// Name stores the constraint name when available.
	Name string

	// Expression stores the check expression when safely available.
	Expression string

	// Raw stores unnormalized check constraint metadata.
	Raw RawMetadata
}

// Index describes database index metadata.
type Index struct {
	// Name stores the index name.
	Name string

	// Columns stores indexed column names when the index is column based.
	Columns []string

	// Unique reports whether the index enforces uniqueness.
	Unique bool

	// Primary reports whether the index backs a primary key.
	Primary bool

	// Expression stores expression index text when safely available.
	Expression string

	// Raw stores unnormalized index metadata.
	Raw RawMetadata
}
