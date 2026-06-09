package connection

import "encoding/json"

// DatabaseType identifies a business-level database connection type in the connection domain.
// It stores the user's intended database family for connection configuration and persistence.
// It is not a direct replacement for internal/dbx/core.DBType, which represents adapter capability.
type DatabaseType string

const (
	// DatabaseTypeMySQL identifies MySQL-compatible business connection configurations.
	DatabaseTypeMySQL DatabaseType = "mysql"

	// DatabaseTypePostgreSQL identifies PostgreSQL business connection configurations.
	DatabaseTypePostgreSQL DatabaseType = "postgres"

	// DatabaseTypeSQLite identifies SQLite business connection configurations reserved by the domain model.
	DatabaseTypeSQLite DatabaseType = "sqlite"

	// DatabaseTypeOracle identifies Oracle business connection configurations reserved by the domain model.
	DatabaseTypeOracle DatabaseType = "oracle"

	// DatabaseTypeSQLServer identifies SQL Server business connection configurations reserved by the domain model.
	DatabaseTypeSQLServer DatabaseType = "sqlserver"

	// DatabaseTypeClickHouse identifies ClickHouse business connection configurations reserved by the domain model.
	DatabaseTypeClickHouse DatabaseType = "clickhouse"

	// DatabaseTypeTiDB identifies TiDB business connection configurations reserved by the domain model.
	DatabaseTypeTiDB DatabaseType = "tidb"

	// DatabaseTypeHive identifies Hive business connection configurations reserved by the domain model.
	DatabaseTypeHive DatabaseType = "hive"
)

// IsKnown reports whether the database type belongs to the connection domain's supported or reserved type set.
// It does not report whether a real database adapter has been implemented for the type.
func (t DatabaseType) IsKnown() bool {
	switch t {
	case DatabaseTypeMySQL,
		DatabaseTypePostgreSQL,
		DatabaseTypeSQLite,
		DatabaseTypeOracle,
		DatabaseTypeSQLServer,
		DatabaseTypeClickHouse,
		DatabaseTypeTiDB,
		DatabaseTypeHive:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the database type is outside the connection domain's supported or reserved type set.
// Callers can use this to explicitly reject unrecognized values after JSON deserialization.
func (t DatabaseType) IsUnknown() bool {
	return !t.IsKnown()
}

// IsPrimarySupported reports whether the database type is a first-phase priority validation target.
// It only expresses the current priority boundary and does not guarantee that real connectivity is available.
func (t DatabaseType) IsPrimarySupported() bool {
	switch t {
	case DatabaseTypeMySQL,
		DatabaseTypePostgreSQL:
		return true
	default:
		return false
	}
}

// RequiresNetworkAddress reports whether the known database type needs host validation for network access.
// Unknown values return false so callers can report an unknown type instead of misleading host errors.
func (t DatabaseType) RequiresNetworkAddress() bool {
	switch t {
	case DatabaseTypeMySQL,
		DatabaseTypePostgreSQL,
		DatabaseTypeOracle,
		DatabaseTypeSQLServer,
		DatabaseTypeClickHouse,
		DatabaseTypeTiDB,
		DatabaseTypeHive:
		return true
	default:
		return false
	}
}

// String returns the stable string representation used for persistence and transport.
// Unknown values are returned unchanged so callers can detect and report them explicitly.
func (t DatabaseType) String() string {
	return string(t)
}

// MarshalJSON serializes the database type as its stable string value.
// Unknown values are preserved as strings so validation can classify them explicitly.
func (t DatabaseType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON restores the database type from its serialized string value.
// Unknown strings are preserved instead of being coerced, allowing validation to reject them explicitly.
func (t *DatabaseType) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*t = DatabaseType(value)
	return nil
}
