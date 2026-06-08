package repository

import (
	"context"

	"github.com/gerdong/loomidbx/internal/storage"
)

// Repositories is the service-facing local data access boundary.
type Repositories interface {
	// StorageDiagnostics returns sanitized local storage initialization diagnostics.
	StorageDiagnostics(ctx context.Context) (storage.StorageDiagnostics, error)

	// UnimplementedBusinessRepository returns the stable error for future business repositories.
	UnimplementedBusinessRepository(ctx context.Context, capability string) error
}

// Factory constructs repository sets from initialized storage state.
type Factory struct {
	diagnostics storage.StorageDiagnostics
}

// NewFactory creates a repository factory from storage diagnostics.
func NewFactory(diagnostics storage.StorageDiagnostics) Factory {
	return Factory{diagnostics: diagnostics}
}

// Repositories returns the repository set for service-layer callers.
func (factory Factory) Repositories() Repositories {
	return storageRepositories{diagnostics: factory.diagnostics}
}

type storageRepositories struct {
	diagnostics storage.StorageDiagnostics
}

// StorageDiagnostics returns sanitized local storage initialization diagnostics.
func (repos storageRepositories) StorageDiagnostics(ctx context.Context) (storage.StorageDiagnostics, error) {
	if err := ctx.Err(); err != nil {
		return storage.StorageDiagnostics{}, err
	}
	if !repos.diagnostics.Ready {
		return storage.StorageDiagnostics{}, NewError(ErrorCodeStorageUnavailable, "storage", "read diagnostics", "storage is not initialized", nil)
	}
	return repos.diagnostics, nil
}

// UnimplementedBusinessRepository returns the stable error for future business repositories.
func (repos storageRepositories) UnimplementedBusinessRepository(ctx context.Context, capability string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return NewError(ErrorCodeNotImplemented, capability, "open business repository", "business repository is not implemented in this spec", nil)
}
