package storage

import "context"

// RunMode identifies the runtime mode passed from the configuration system.
type RunMode string

const (
	// ModeDesktop identifies normal desktop execution with the user's configured data directory.
	ModeDesktop RunMode = "desktop"

	// ModeDevelopment identifies development execution with an explicitly configured data directory.
	ModeDevelopment RunMode = "development"

	// ModeTest identifies test execution and requires callers to provide an isolated data directory.
	ModeTest RunMode = "test"
)

// PathSource identifies where storage path information came from.
type PathSource string

const (
	// PathSourceConfig means storage paths were derived only from the resolved configuration data directory.
	PathSourceConfig PathSource = "config"
)

// StorageConfig contains the already-resolved storage startup input from the config module.
type StorageConfig struct {
	// DataDir is the absolute data directory resolved by the configuration system.
	DataDir string

	// Mode is the runtime mode resolved by the configuration system.
	Mode RunMode
}

// ResolvedStoragePaths contains the stable file and directory layout under the configured data directory.
type ResolvedStoragePaths struct {
	// RootDir is the configured local storage root directory.
	RootDir string

	// DatabaseFile is the main local structured business data file.
	DatabaseFile string

	// MigrationDir is the directory reserved for migration state and downstream migration files.
	MigrationDir string

	// TempDir is the directory reserved for temporary local storage files.
	TempDir string

	// BackupDir is the directory reserved for future backup or export files.
	BackupDir string

	// Source records that the layout came from the resolved config data directory.
	Source PathSource
}

// StorageDiagnostics is the sanitized runtime view returned after storage initialization.
type StorageDiagnostics struct {
	// DataDir is the configured local storage root directory.
	DataDir string

	// DatabaseFile is the main local structured business data file.
	DatabaseFile string

	// MigrationVersion is the latest successfully applied migration version.
	MigrationVersion string

	// PendingMigrations is the number of migrations still pending after initialization.
	PendingMigrations int

	// AppliedMigrations is the total number of successfully applied migrations known to storage.
	AppliedMigrations int

	// BusinessTablesCreated reports how many business tables this spec created; it must remain zero.
	BusinessTablesCreated int

	// Ready reports whether storage initialization completed successfully.
	Ready bool
}

// LayoutResolver resolves and validates storage paths from configuration output.
type LayoutResolver interface {
	// Resolve returns stable storage paths for the provided config or a typed path error.
	Resolve(ctx context.Context, config StorageConfig) (ResolvedStoragePaths, error)
}

// Bootstrapper initializes local storage without depending on frontend UI.
type Bootstrapper interface {
	// Initialize prepares directories, opens the data file, applies base migrations, and returns diagnostics.
	Initialize(ctx context.Context, config StorageConfig) (StorageDiagnostics, error)
}
