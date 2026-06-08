package storage_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gerdong/loomidbx/internal/repository"
	"github.com/gerdong/loomidbx/internal/storage"
)

func TestEndToEndInitializeAndRepositoryDiagnostics(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	bootstrapper := storage.NewBootstrapper()

	diagnostics, initErr := bootstrapper.Initialize(ctx, storage.StorageConfig{
		DataDir: root,
		Mode:    storage.ModeTest,
	})
	if initErr != nil {
		t.Fatalf("Initialize() error = %v, want nil", initErr)
	}

	unit := repository.NewUnitOfWork(repository.NewFactory(diagnostics))
	err := unit.Do(ctx, func(ctx context.Context, repos repository.Repositories) error {
		got, repoErr := repos.StorageDiagnostics(ctx)
		if repoErr != nil {
			return repoErr
		}
		if got != diagnostics {
			t.Fatalf("repository diagnostics = %+v, want bootstrap diagnostics %+v", got, diagnostics)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("UnitOfWork.Do() error = %v, want nil", err)
	}
}

func TestEndToEndInitializationFailureUsesStableSanitizedError(t *testing.T) {
	ctx := context.Background()
	bootstrapper := storage.NewBootstrapper()

	_, err := bootstrapper.Initialize(ctx, storage.StorageConfig{
		DataDir: "relative/password-secret-token",
		Mode:    storage.ModeTest,
	})

	var storageErr *storage.StorageError
	if !errors.As(err, &storageErr) {
		t.Fatalf("Initialize() error = %T %[1]v, want StorageError", err)
	}
	if storageErr.Code != storage.ErrorCodePathInvalid {
		t.Fatalf("Code = %q, want %q", storageErr.Code, storage.ErrorCodePathInvalid)
	}
	if strings.Contains(storageErr.Error(), "password-secret-token") {
		t.Fatalf("Error() = %q, want sensitive value redacted", storageErr.Error())
	}
}
