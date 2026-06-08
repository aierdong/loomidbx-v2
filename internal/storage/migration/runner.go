package migration

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Connection is the minimal local structured data connection required by the migration runner.
type Connection interface {
	// EnsureMigrationStore prepares the migration record storage.
	EnsureMigrationStore(ctx context.Context) error

	// AppliedRecords returns successfully applied migration records by version.
	AppliedRecords(ctx context.Context) (map[string]Record, error)

	// Execute runs a migration body.
	Execute(ctx context.Context, statement string) error

	// RecordSuccess records a successfully applied migration.
	RecordSuccess(ctx context.Context, record Record) error
}

// Result summarizes migration execution state for diagnostics.
type Result struct {
	// CurrentVersion is the latest successfully applied migration version.
	CurrentVersion string

	// AppliedCount is the number of migrations applied during this run.
	AppliedCount int

	// TotalApplied is the total number of applied migration records after this run.
	TotalApplied int

	// PendingCount is the number of migrations still pending after this run.
	PendingCount int
}

// Runner applies registered migrations in deterministic order.
type Runner struct {
	registry Registry
}

// NewRunner creates a migration runner from a registry.
func NewRunner(registry Registry) Runner {
	return Runner{registry: registry}
}

// Apply applies pending migrations and returns sanitized migration state.
func (runner Runner) Apply(ctx context.Context, conn Connection) (Result, error) {
	if conn == nil {
		return Result{}, NewMigrationError("", "apply migrations", "migration connection is required", nil)
	}
	registry := runner.registry
	if registry == nil {
		registry = DefaultRegistry()
	}
	migrations, err := registry.List()
	if err != nil {
		return Result{}, err
	}
	if err := conn.EnsureMigrationStore(ctx); err != nil {
		return Result{}, NewMigrationError("", "prepare migration store", "could not prepare migration records", err)
	}
	records, err := conn.AppliedRecords(ctx)
	if err != nil {
		return Result{}, NewMigrationError("", "read migration records", "could not read migration records", err)
	}
	appliedThisRun := 0
	for _, item := range migrations {
		if _, ok := records[item.Version]; ok {
			continue
		}
		if err := conn.Execute(ctx, item.Up); err != nil {
			return Result{}, NewMigrationError(item.Version, "execute migration", "migration execution failed", err)
		}
		record := Record{
			Version:   item.Version,
			Name:      item.Name,
			Checksum:  Checksum(item.Up),
			AppliedAt: time.Now().Unix(),
		}
		if err := conn.RecordSuccess(ctx, record); err != nil {
			return Result{}, NewMigrationError(item.Version, "record migration", "could not record migration success", err)
		}
		records[item.Version] = record
		appliedThisRun++
	}
	current := latestVersion(records)
	pending := 0
	for _, item := range migrations {
		if _, ok := records[item.Version]; !ok {
			pending++
		}
	}
	return Result{CurrentVersion: current, AppliedCount: appliedThisRun, TotalApplied: len(records), PendingCount: pending}, nil
}

func latestVersion(records map[string]Record) string {
	latest := ""
	for version := range records {
		if version > latest {
			latest = version
		}
	}
	return latest
}

// SQLiteConnection is a SQLite-backed local migration connection.
type SQLiteConnection struct {
	// Path is the local SQLite database file path.
	Path string

	db *sql.DB
}

// OpenFileConnection opens or creates the local SQLite structured data file.
func OpenFileConnection(path string) (*SQLiteConnection, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &SQLiteConnection{Path: path, db: db}, nil
}

// Close releases the SQLite connection resources.
func (conn *SQLiteConnection) Close() error {
	if conn == nil || conn.db == nil {
		return nil
	}
	return conn.db.Close()
}

// EnsureMigrationStore prepares the migration record storage.
func (conn *SQLiteConnection) EnsureMigrationStore(ctx context.Context) error {
	_, err := conn.db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, name TEXT NOT NULL, checksum TEXT NOT NULL, applied_at INTEGER NOT NULL)")
	return err
}

// AppliedRecords returns successfully applied migration records by version.
func (conn *SQLiteConnection) AppliedRecords(ctx context.Context) (map[string]Record, error) {
	rows, err := conn.db.QueryContext(ctx, "SELECT version, name, checksum, applied_at FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := map[string]Record{}
	for rows.Next() {
		var record Record
		if err := rows.Scan(&record.Version, &record.Name, &record.Checksum, &record.AppliedAt); err != nil {
			return nil, err
		}
		records[record.Version] = record
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

// Execute runs a migration body inside SQLite.
func (conn *SQLiteConnection) Execute(ctx context.Context, statement string) error {
	_, err := conn.db.ExecContext(ctx, statement)
	return err
}

// RecordSuccess records a successfully applied migration.
func (conn *SQLiteConnection) RecordSuccess(ctx context.Context, record Record) error {
	_, err := conn.db.ExecContext(ctx, "INSERT INTO schema_migrations(version, name, checksum, applied_at) VALUES (?, ?, ?, ?)", record.Version, record.Name, record.Checksum, record.AppliedAt)
	return err
}

// MemoryConnection is a deterministic in-memory connection for migration tests and repository substitutes.
type MemoryConnection struct {
	mu      sync.Mutex
	records map[string]Record
	execs   map[string]int
}

// NewMemoryConnection returns an empty migration connection test double.
func NewMemoryConnection() *MemoryConnection {
	return &MemoryConnection{records: map[string]Record{}, execs: map[string]int{}}
}

// EnsureMigrationStore prepares the migration record storage.
func (conn *MemoryConnection) EnsureMigrationStore(ctx context.Context) error {
	return ctx.Err()
}

// AppliedRecords returns a copy of successful migration records by version.
func (conn *MemoryConnection) AppliedRecords(ctx context.Context) (map[string]Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	conn.mu.Lock()
	defer conn.mu.Unlock()
	copied := map[string]Record{}
	for version, record := range conn.records {
		copied[version] = record
	}
	return copied, nil
}

// Execute records a migration execution or returns a deterministic failure when the body starts with FAIL.
func (conn *MemoryConnection) Execute(ctx context.Context, statement string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.execs[statement]++
	if strings.HasPrefix(statement, "FAIL") {
		return fmt.Errorf("migration statement failed")
	}
	return nil
}

// RecordSuccess records a successfully applied migration.
func (conn *MemoryConnection) RecordSuccess(ctx context.Context, record Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.records[record.Version] = record
	return nil
}

// HasApplied reports whether a migration version has been recorded as successful.
func (conn *MemoryConnection) HasApplied(version string) bool {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	_, ok := conn.records[version]
	return ok
}

// ExecCount reports how many times a migration statement was executed.
func (conn *MemoryConnection) ExecCount(statement string) int {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	return conn.execs[statement]
}
