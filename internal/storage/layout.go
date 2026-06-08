package storage

import (
	"context"
	"os"
	"path/filepath"
)

// DefaultLayoutResolver resolves the fixed storage layout under the configured data directory.
type DefaultLayoutResolver struct{}

// NewLayoutResolver returns the default config-driven layout resolver.
func NewLayoutResolver() DefaultLayoutResolver {
	return DefaultLayoutResolver{}
}

// Resolve returns stable storage paths for the provided config or a typed path error.
func (resolver DefaultLayoutResolver) Resolve(ctx context.Context, config StorageConfig) (ResolvedStoragePaths, error) {
	if err := ctx.Err(); err != nil {
		return ResolvedStoragePaths{}, NewError(ErrorCodePathInvalid, "context", "resolve layout", "context canceled", err)
	}
	if config.DataDir == "" {
		return ResolvedStoragePaths{}, NewError(ErrorCodePathInvalid, "dataDir", "resolve layout", "data directory is required", nil)
	}
	if !filepath.IsAbs(config.DataDir) {
		return ResolvedStoragePaths{}, NewError(ErrorCodePathInvalid, "dataDir", "resolve layout", "data directory must be absolute", nil)
	}
	root := filepath.Clean(config.DataDir)
	paths := ResolvedStoragePaths{
		RootDir:      root,
		DatabaseFile: filepath.Join(root, "loomidbx.db"),
		MigrationDir: filepath.Join(root, "migrations"),
		TempDir:      filepath.Join(root, "tmp"),
		BackupDir:    filepath.Join(root, "backups"),
		Source:       PathSourceConfig,
	}
	if err := ensureDirectory(paths.RootDir); err != nil {
		return ResolvedStoragePaths{}, NewError(ErrorCodePathInvalid, "dataDir", "resolve layout", "data directory is not writable", err)
	}
	return paths, nil
}

func ensureDirectory(path string) error {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return err
	}
	probe, err := os.CreateTemp(path, ".loomidbx-write-test-*")
	if err != nil {
		return err
	}
	probeName := probe.Name()
	if closeErr := probe.Close(); closeErr != nil {
		_ = os.Remove(probeName)
		return closeErr
	}
	return os.Remove(probeName)
}
