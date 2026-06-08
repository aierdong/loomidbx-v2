package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gerdong/loomidbx/internal/repository"
	"github.com/gerdong/loomidbx/internal/storage"
)

func TestFactoryExposesStorageDiagnosticsWithoutFilePathAccess(t *testing.T) {
	ctx := context.Background()
	diagnostics := storage.StorageDiagnostics{
		DataDir:          "/isolated/data",
		DatabaseFile:     "/isolated/data/loomidbx.db",
		MigrationVersion: "000001",
		Ready:            true,
	}
	factory := repository.NewFactory(diagnostics)

	got, err := factory.Repositories().StorageDiagnostics(ctx)

	if err != nil {
		t.Fatalf("StorageDiagnostics() error = %v, want nil", err)
	}
	if got != diagnostics {
		t.Fatalf("StorageDiagnostics() = %+v, want %+v", got, diagnostics)
	}
}

func TestUnitOfWorkProvidesRepositoriesAndPropagatesCallbackErrors(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("service validation failed")
	unit := repository.NewUnitOfWork(repository.NewFactory(storage.StorageDiagnostics{Ready: true}))

	err := unit.Do(ctx, func(ctx context.Context, repos repository.Repositories) error {
		if _, diagErr := repos.StorageDiagnostics(ctx); diagErr != nil {
			t.Fatalf("StorageDiagnostics() error = %v, want nil", diagErr)
		}
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("Do() error = %v, want callback error %v", err, wantErr)
	}
}

func TestUnavailableBusinessRepositoryReturnsNotImplemented(t *testing.T) {
	ctx := context.Background()
	factory := repository.NewFactory(storage.StorageDiagnostics{Ready: true})

	err := factory.Repositories().UnimplementedBusinessRepository(ctx, "project")

	var repoErr *repository.RepositoryError
	if !errors.As(err, &repoErr) {
		t.Fatalf("UnimplementedBusinessRepository() error = %T %[1]v, want RepositoryError", err)
	}
	if repoErr.Code != repository.ErrorCodeNotImplemented {
		t.Fatalf("Code = %q, want %q", repoErr.Code, repository.ErrorCodeNotImplemented)
	}
	if repoErr.Capability != "project" {
		t.Fatalf("Capability = %q, want project", repoErr.Capability)
	}
}

func TestUnavailableStorageReturnsStorageUnavailable(t *testing.T) {
	ctx := context.Background()
	factory := repository.NewFactory(storage.StorageDiagnostics{Ready: false})

	_, err := factory.Repositories().StorageDiagnostics(ctx)

	var repoErr *repository.RepositoryError
	if !errors.As(err, &repoErr) {
		t.Fatalf("StorageDiagnostics() error = %T %[1]v, want RepositoryError", err)
	}
	if repoErr.Code != repository.ErrorCodeStorageUnavailable {
		t.Fatalf("Code = %q, want %q", repoErr.Code, repository.ErrorCodeStorageUnavailable)
	}
}
