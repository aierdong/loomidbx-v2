package capability

// Capabilities describes database behavior that downstream services can negotiate at runtime.
type Capabilities struct {
	// SupportsTransactions reports whether the database can execute transactional work.
	SupportsTransactions bool

	// SupportsSavepoints reports whether nested savepoint operations are available.
	SupportsSavepoints bool

	// SupportsForeignKeys reports whether the database enforces foreign key constraints.
	SupportsForeignKeys bool

	// SupportsDeferredConstraints reports whether constraint checks can be deferred.
	SupportsDeferredConstraints bool

	// SupportsBatchInsert reports whether multi-row insert statements are supported.
	SupportsBatchInsert bool

	// SupportsBulkLoad reports whether dedicated bulk-load primitives may exist.
	SupportsBulkLoad bool

	// SupportsReturning reports whether write statements can return changed rows.
	SupportsReturning bool

	// SupportsUpsert reports whether insert-or-update semantics are available.
	SupportsUpsert bool

	// SupportsCatalogs reports whether catalog-level names are meaningful.
	SupportsCatalogs bool

	// SupportsSchemas reports whether schema or namespace-level names are meaningful.
	SupportsSchemas bool

	// SupportsJSON reports whether JSON values are supported natively or equivalently.
	SupportsJSON bool

	// SupportsArrays reports whether array values are supported natively or equivalently.
	SupportsArrays bool

	// SupportsUUID reports whether UUID values are supported natively or equivalently.
	SupportsUUID bool

	// SupportsEnums reports whether enum values are supported natively or equivalently.
	SupportsEnums bool

	// SupportsGeneratedColumns reports whether generated columns are supported.
	SupportsGeneratedColumns bool

	// SupportsIdentityColumns reports whether identity columns are supported.
	SupportsIdentityColumns bool

	// MaxIdentifierLength is the maximum identifier length when the adapter can state one.
	MaxIdentifierLength int

	// MaxParameters is the maximum bind parameter count when constrained.
	MaxParameters int

	// MaxBatchRows is the maximum rows per batch insert when constrained.
	MaxBatchRows int
}

// MySQLExample returns a documentation-oriented capability sample for tests and strategy examples.
func MySQLExample() Capabilities {
	return Capabilities{
		SupportsTransactions:     true,
		SupportsSavepoints:       true,
		SupportsForeignKeys:      true,
		SupportsBatchInsert:      true,
		SupportsBulkLoad:         true,
		SupportsUpsert:           true,
		SupportsSchemas:          true,
		SupportsJSON:             true,
		SupportsGeneratedColumns: true,
		SupportsIdentityColumns:  true,
		MaxIdentifierLength:      64,
	}
}

// PostgreSQLExample returns a documentation-oriented capability sample for tests and strategy examples.
func PostgreSQLExample() Capabilities {
	return Capabilities{
		SupportsTransactions:        true,
		SupportsSavepoints:          true,
		SupportsForeignKeys:         true,
		SupportsDeferredConstraints: true,
		SupportsBatchInsert:         true,
		SupportsBulkLoad:            true,
		SupportsReturning:           true,
		SupportsUpsert:              true,
		SupportsCatalogs:            true,
		SupportsSchemas:             true,
		SupportsJSON:                true,
		SupportsArrays:              true,
		SupportsUUID:                true,
		SupportsEnums:               true,
		SupportsGeneratedColumns:    true,
		SupportsIdentityColumns:     true,
		MaxIdentifierLength:         63,
	}
}
