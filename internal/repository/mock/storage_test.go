package mock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gerdong/loomidbx/internal/repository"
	repositorymock "github.com/gerdong/loomidbx/internal/repository/mock"
	"github.com/gerdong/loomidbx/internal/storage"
)

func TestMockRepositoriesReturnConfiguredDiagnostics(t *testing.T) {
	ctx := context.Background()
	diagnostics := storage.StorageDiagnostics{Ready: true, MigrationVersion: "000001"}
	repos := repositorymock.NewRepositories(diagnostics)

	got, err := repos.StorageDiagnostics(ctx)

	if err != nil {
		t.Fatalf("StorageDiagnostics() error = %v, want nil", err)
	}
	if got != diagnostics {
		t.Fatalf("StorageDiagnostics() = %+v, want %+v", got, diagnostics)
	}
}

func TestMockRepositoriesPreserveUnimplementedErrorSemantics(t *testing.T) {
	ctx := context.Background()
	repos := repositorymock.NewRepositories(storage.StorageDiagnostics{Ready: true})

	err := repos.UnimplementedBusinessRepository(ctx, "connection")

	var repoErr *repository.RepositoryError
	if !errors.As(err, &repoErr) {
		t.Fatalf("UnimplementedBusinessRepository() error = %T %[1]v, want RepositoryError", err)
	}
	if repoErr.Code != repository.ErrorCodeNotImplemented {
		t.Fatalf("Code = %q, want %q", repoErr.Code, repository.ErrorCodeNotImplemented)
	}
}
