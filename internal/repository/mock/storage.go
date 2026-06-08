package mock

import (
	"context"

	"github.com/gerdong/loomidbx/internal/repository"
	"github.com/gerdong/loomidbx/internal/storage"
)

// Repositories is a service-test double that preserves repository error semantics.
type Repositories struct {
	// Diagnostics is the storage diagnostics returned by the mock repository set.
	Diagnostics storage.StorageDiagnostics
}

// NewRepositories creates a mock repository set for service unit tests.
func NewRepositories(diagnostics storage.StorageDiagnostics) Repositories {
	return Repositories{Diagnostics: diagnostics}
}

// StorageDiagnostics returns the configured diagnostics without touching SQLite files.
func (repos Repositories) StorageDiagnostics(ctx context.Context) (storage.StorageDiagnostics, error) {
	if err := ctx.Err(); err != nil {
		return storage.StorageDiagnostics{}, err
	}
	if !repos.Diagnostics.Ready {
		return storage.StorageDiagnostics{}, repository.NewError(repository.ErrorCodeStorageUnavailable, "storage", "read diagnostics", "storage is not initialized", nil)
	}
	return repos.Diagnostics, nil
}

// UnimplementedBusinessRepository returns the same stable error as real repositories.
func (repos Repositories) UnimplementedBusinessRepository(ctx context.Context, capability string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return repository.NewError(repository.ErrorCodeNotImplemented, capability, "open business repository", "business repository is not implemented in this spec", nil)
}
