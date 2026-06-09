package connection

// DatabaseType is the connection-domain database type selected by users and persisted as a stable string.
type DatabaseType string

const (
	// DatabaseTypeMySQL identifies MySQL-compatible business connection configurations.
	DatabaseTypeMySQL DatabaseType = "mysql"

	// DatabaseTypePostgres identifies PostgreSQL-compatible business connection configurations.
	DatabaseTypePostgres DatabaseType = "postgres"

	// DatabaseTypeSQLite identifies SQLite business connection configurations reserved for local-file databases.
	DatabaseTypeSQLite DatabaseType = "sqlite"

	// DatabaseTypeOracle identifies Oracle business connection configurations reserved for future adapter support.
	DatabaseTypeOracle DatabaseType = "oracle"

	// DatabaseTypeSQLServer identifies SQL Server business connection configurations reserved for future adapter support.
	DatabaseTypeSQLServer DatabaseType = "sqlserver"

	// DatabaseTypeClickHouse identifies ClickHouse business connection configurations reserved for future adapter support.
	DatabaseTypeClickHouse DatabaseType = "clickhouse"

	// DatabaseTypeTiDB identifies TiDB business connection configurations reserved for future adapter support.
	DatabaseTypeTiDB DatabaseType = "tidb"

	// DatabaseTypeHive identifies Hive business connection configurations reserved for future adapter support.
	DatabaseTypeHive DatabaseType = "hive"
)
