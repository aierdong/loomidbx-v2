package migration

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const errorCodeMigrationFailed = "STORAGE_MIGRATION_FAILED"

// Migration describes one ordered storage schema migration.
type Migration struct {
	// Version is a six-digit increasing migration number such as 000001.
	Version string

	// Name is a semantic non-sensitive migration name.
	Name string

	// Up contains the migration operation body.
	Up string
}

// Record stores successful migration metadata.
type Record struct {
	// Version is the successfully applied migration version.
	Version string

	// Name is the migration name recorded with the version.
	Name string

	// Checksum is the SHA-256 checksum of the migration body.
	Checksum string

	// AppliedAt is the Unix timestamp in seconds when the migration succeeded.
	AppliedAt int64
}

// Registry returns migrations in deterministic execution order.
type Registry interface {
	// List returns sorted, validated migration definitions.
	List() ([]Migration, error)
}

// StaticRegistry stores an in-memory migration list for default and test use.
type StaticRegistry struct {
	migrations []Migration
}

// NewRegistry creates a static registry from migration definitions.
func NewRegistry(migrations []Migration) StaticRegistry {
	copied := append([]Migration(nil), migrations...)
	return StaticRegistry{migrations: copied}
}

// DefaultRegistry returns the base metadata migration registry for this spec.
func DefaultRegistry() StaticRegistry {
	return NewRegistry([]Migration{{
		Version: "000001",
		Name:    "storage_metadata",
		Up:      "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, name TEXT NOT NULL, checksum TEXT NOT NULL, applied_at INTEGER NOT NULL)",
	}})
}

// List returns sorted, validated migration definitions.
func (registry StaticRegistry) List() ([]Migration, error) {
	migrations := append([]Migration(nil), registry.migrations...)
	for _, item := range migrations {
		if err := validateMigration(item); err != nil {
			return nil, err
		}
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	for index := 1; index < len(migrations); index++ {
		if migrations[index].Version == migrations[index-1].Version {
			return nil, NewMigrationError(migrations[index].Version, "validate registry", "duplicate migration version", nil)
		}
	}
	return migrations, nil
}

// Checksum returns the stable checksum for a migration body.
func Checksum(up string) string {
	sum := sha256.Sum256([]byte(up))
	return hex.EncodeToString(sum[:])
}

func validateMigration(item Migration) error {
	if len(item.Version) != 6 {
		return NewMigrationError(item.Version, "validate registry", "migration version must use six digits", nil)
	}
	if _, err := strconv.Atoi(item.Version); err != nil {
		return NewMigrationError(item.Version, "validate registry", "migration version must be numeric", err)
	}
	if item.Name == "" || strings.Contains(item.Name, " ") {
		return NewMigrationError(item.Version, "validate registry", "migration name is required", nil)
	}
	if item.Up == "" {
		return NewMigrationError(item.Version, "validate registry", "migration body is required", nil)
	}
	return nil
}

// MigrationError carries sanitized migration failure context without depending on parent storage packages.
type MigrationError struct {
	// Code is the stable storage migration error category.
	Code string

	// Field is the migration field related to the failure.
	Field string

	// Operation is the migration operation that failed.
	Operation string

	// MigrationVersion is the failed migration version when known.
	MigrationVersion string

	// Reason is a sanitized human-readable reason.
	Reason string

	// Err is the wrapped cause for Go error chaining.
	Err error
}

// Error returns a sanitized migration error message.
func (err *MigrationError) Error() string {
	if err == nil {
		return "storage migration error"
	}
	message := "storage " + err.Code
	if err.Operation != "" {
		message += " during " + err.Operation
	}
	if err.Field != "" {
		message += " on " + err.Field
	}
	if err.MigrationVersion != "" {
		message += " for migration " + err.MigrationVersion
	}
	if err.Reason != "" {
		message += ": " + err.Reason
	}
	return message
}

// Unwrap returns the wrapped non-sensitive cause.
func (err *MigrationError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

// NewMigrationError constructs a typed migration failure error.
func NewMigrationError(version string, operation string, reason string, cause error) *MigrationError {
	var wrapped error
	if cause != nil {
		wrapped = fmt.Errorf("[redacted]")
	}
	return &MigrationError{Code: errorCodeMigrationFailed, Field: "migration", Operation: operation, MigrationVersion: version, Reason: "[redacted]", Err: wrapped}
}
