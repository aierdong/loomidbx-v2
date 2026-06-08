package core

import "fmt"

// ErrorKind identifies a DBX error category for errors.Is matching.
type ErrorKind string

const (
	// ErrorUnsupportedDatabase reports that no adapter exists for a database type.
	ErrorUnsupportedDatabase ErrorKind = "unsupported_database"

	// ErrorInvalidAdapter reports a nil or malformed adapter registration.
	ErrorInvalidAdapter ErrorKind = "invalid_adapter"

	// ErrorDuplicateAdapter reports a deterministic duplicate adapter registration failure.
	ErrorDuplicateAdapter ErrorKind = "duplicate_adapter"

	// ErrorInvalidConnectionConfig reports invalid connection configuration at the adapter boundary.
	ErrorInvalidConnectionConfig ErrorKind = "invalid_connection_config"

	// ErrorUnsupportedDialectOperation reports a dialect operation that cannot be built.
	ErrorUnsupportedDialectOperation ErrorKind = "unsupported_dialect_operation"

	// ErrorIntrospectionFailed reports metadata scanning failure.
	ErrorIntrospectionFailed ErrorKind = "introspection_failed"

	// ErrorTypeMappingFailed reports native-to-logical type mapping failure.
	ErrorTypeMappingFailed ErrorKind = "type_mapping_failed"
)

// DBXError carries a typed category and non-sensitive context.
type DBXError struct {
	// Kind identifies the stable error category.
	Kind ErrorKind

	// DBType stores the database type related to the failure when known.
	DBType DBType

	// Operation stores the failed operation name when useful.
	Operation string

	// Object stores a schema, table, column, or adapter name when useful.
	Object string

	// Err stores the wrapped cause.
	Err error
}

// Error returns a sanitized error message with non-sensitive context only.
func (e *DBXError) Error() string {
	if e == nil {
		return "dbx error"
	}
	msg := "dbx " + string(e.Kind)
	if e.DBType != "" {
		msg += " for database " + string(e.DBType)
	}
	if e.Operation != "" {
		msg += " during " + e.Operation
	}
	if e.Object != "" {
		msg += " on " + e.Object
	}
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}

// Unwrap returns the wrapped cause for standard Go error chaining.
func (e *DBXError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is matches DBX errors by category for errors.Is.
func (e *DBXError) Is(target error) bool {
	targetErr, ok := target.(*DBXError)
	if !ok || e == nil || targetErr == nil {
		return false
	}
	return targetErr.Kind == "" || e.Kind == targetErr.Kind
}

// NewError constructs a typed DBX error with sanitized context.
func NewError(kind ErrorKind, dbType DBType, operation string, object string, err error) *DBXError {
	return &DBXError{Kind: kind, DBType: dbType, Operation: operation, Object: object, Err: err}
}

// UnsupportedDatabaseError constructs a typed missing-adapter error.
func UnsupportedDatabaseError(dbType DBType) *DBXError {
	return NewError(ErrorUnsupportedDatabase, dbType, "lookup adapter", "", nil)
}

// InvalidAdapterError constructs a typed adapter validation error.
func InvalidAdapterError(reason string) *DBXError {
	return NewError(ErrorInvalidAdapter, "", "register adapter", reason, nil)
}

// DuplicateAdapterError constructs a typed duplicate adapter registration error.
func DuplicateAdapterError(dbType DBType) *DBXError {
	return NewError(ErrorDuplicateAdapter, dbType, "register adapter", "", nil)
}

// UnsupportedDialectOperationError constructs a typed unsupported dialect operation error.
func UnsupportedDialectOperationError(operation string) *DBXError {
	return NewError(ErrorUnsupportedDialectOperation, "", operation, "", nil)
}

// WrapIntrospectionError constructs a typed introspection failure error.
func WrapIntrospectionError(dbType DBType, operation string, object string, err error) *DBXError {
	if err == nil {
		err = fmt.Errorf("introspection failed")
	}
	return NewError(ErrorIntrospectionFailed, dbType, operation, object, err)
}

// WrapTypeMappingError constructs a typed type mapping failure error.
func WrapTypeMappingError(dbType DBType, object string, err error) *DBXError {
	if err == nil {
		err = fmt.Errorf("type mapping failed")
	}
	return NewError(ErrorTypeMappingFailed, dbType, "map native type", object, err)
}
