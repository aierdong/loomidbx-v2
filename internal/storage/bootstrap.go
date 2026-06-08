package storage

import (
	"context"
	"errors"
	"os"

	"github.com/gerdong/loomidbx/internal/storage/migration"
)

// DefaultBootstrapper initializes storage using the default layout resolver and base migrations.
type DefaultBootstrapper struct {
	// Resolver resolves the local storage file layout.
	Resolver LayoutResolver

	// Runner applies pending migrations to the local data file.
	Runner MigrationRunner
}

// MigrationRunner is the migration execution boundary used by the storage bootstrapper.
type MigrationRunner interface {
	// Apply applies pending migrations and returns sanitized migration state.
	Apply(ctx context.Context, conn migration.Connection) (migration.Result, error)
}

// NewBootstrapper returns the default storage bootstrapper.
func NewBootstrapper() DefaultBootstrapper {
	return DefaultBootstrapper{
		Resolver: NewLayoutResolver(),
		Runner:   migration.NewRunner(migration.DefaultRegistry()),
	}
}

// Initialize prepares directories, opens the data file, applies base migrations, and returns diagnostics.
func (bootstrapper DefaultBootstrapper) Initialize(ctx context.Context, config StorageConfig) (StorageDiagnostics, error) {
	resolver := bootstrapper.Resolver
	if resolver == nil {
		resolver = NewLayoutResolver()
	}
	paths, err := resolver.Resolve(ctx, config)
	if err != nil {
		return StorageDiagnostics{}, err
	}
	for _, dir := range []string{paths.MigrationDir, paths.TempDir, paths.BackupDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return StorageDiagnostics{}, NewError(ErrorCodePathInvalid, "layout", "create layout directory", "directory is not writable", err)
		}
	}
	conn, err := migration.OpenFileConnection(paths.DatabaseFile)
	if err != nil {
		return StorageDiagnostics{}, NewError(ErrorCodeOpenFailed, "databaseFile", "open database", "database file could not be opened", err)
	}
	defer conn.Close()

	runner := bootstrapper.Runner
	if runner == nil {
		runner = migration.NewRunner(migration.DefaultRegistry())
	}
	result, err := runner.Apply(ctx, conn)
	if err != nil {
		return StorageDiagnostics{}, convertMigrationError(err)
	}
	return StorageDiagnostics{
		DataDir:               paths.RootDir,
		DatabaseFile:          paths.DatabaseFile,
		MigrationVersion:      result.CurrentVersion,
		PendingMigrations:     result.PendingCount,
		AppliedMigrations:     result.TotalApplied,
		BusinessTablesCreated: 0,
		Ready:                 true,
	}, nil
}

func convertMigrationError(err error) error {
	var migrationErr *migration.MigrationError
	if errors.As(err, &migrationErr) {
		return NewStorageError(migrationErr.Code, migrationErr.Field, migrationErr.Operation, migrationErr.MigrationVersion, migrationErr.Reason, migrationErr.Err)
	}
	return err
}
