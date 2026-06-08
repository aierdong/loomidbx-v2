package storage_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gerdong/loomidbx/internal/storage"
)

func TestResolveLayoutUsesConfigDataDirAsOnlyInput(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	resolver := storage.NewLayoutResolver()

	paths, err := resolver.Resolve(ctx, storage.StorageConfig{
		DataDir: root,
		Mode:    storage.ModeTest,
	})

	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	assertPathEqual(t, paths.RootDir, root)
	assertPathEqual(t, paths.DatabaseFile, filepath.Join(root, "loomidbx.db"))
	assertPathEqual(t, paths.MigrationDir, filepath.Join(root, "migrations"))
	assertPathEqual(t, paths.TempDir, filepath.Join(root, "tmp"))
	assertPathEqual(t, paths.BackupDir, filepath.Join(root, "backups"))
	if paths.Source != storage.PathSourceConfig {
		t.Fatalf("Source = %q, want %q", paths.Source, storage.PathSourceConfig)
	}
}

func TestResolveLayoutRejectsRelativeDataDirWithStableError(t *testing.T) {
	ctx := context.Background()
	resolver := storage.NewLayoutResolver()

	_, err := resolver.Resolve(ctx, storage.StorageConfig{
		DataDir: filepath.Join("relative", "data"),
		Mode:    storage.ModeTest,
	})

	var storageErr *storage.StorageError
	if !errors.As(err, &storageErr) {
		t.Fatalf("Resolve() error = %T %[1]v, want StorageError", err)
	}
	if storageErr.Code != storage.ErrorCodePathInvalid {
		t.Fatalf("Code = %q, want %q", storageErr.Code, storage.ErrorCodePathInvalid)
	}
	if storageErr.Field != "dataDir" {
		t.Fatalf("Field = %q, want dataDir", storageErr.Field)
	}
	if strings.Contains(storageErr.Error(), "relative/data") || strings.Contains(storageErr.Error(), "relative\\data") {
		t.Fatalf("Error() = %q, want path value redacted", storageErr.Error())
	}
}

func TestInitializeCreatesIsolatedLayoutDatabaseAndDiagnostics(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	bootstrapper := storage.NewBootstrapper()

	diagnostics, err := bootstrapper.Initialize(ctx, storage.StorageConfig{
		DataDir: root,
		Mode:    storage.ModeTest,
	})

	if err != nil {
		t.Fatalf("Initialize() error = %v, want nil", err)
	}
	if !diagnostics.Ready {
		t.Fatalf("Ready = false, want true: %+v", diagnostics)
	}
	assertPathEqual(t, diagnostics.DataDir, root)
	assertPathEqual(t, diagnostics.DatabaseFile, filepath.Join(root, "loomidbx.db"))
	if diagnostics.MigrationVersion != "000001" {
		t.Fatalf("MigrationVersion = %q, want 000001", diagnostics.MigrationVersion)
	}
	if diagnostics.PendingMigrations != 0 || diagnostics.AppliedMigrations != 1 {
		t.Fatalf("migration diagnostics = %+v, want one applied and none pending", diagnostics)
	}
	for _, dir := range []string{root, filepath.Join(root, "migrations"), filepath.Join(root, "tmp"), filepath.Join(root, "backups")} {
		info, statErr := os.Stat(dir)
		if statErr != nil {
			t.Fatalf("Stat(%q) error = %v", dir, statErr)
		}
		if !info.IsDir() {
			t.Fatalf("%q is not a directory", dir)
		}
	}
	if info, statErr := os.Stat(diagnostics.DatabaseFile); statErr != nil || info.IsDir() {
		t.Fatalf("database file stat = (%+v, %v), want file", info, statErr)
	}
	if !strings.HasPrefix(filepath.Clean(diagnostics.DatabaseFile), filepath.Clean(root)) {
		t.Fatalf("database file %q is outside isolated test root %q", diagnostics.DatabaseFile, root)
	}
}

func TestInitializeIsIdempotentAndDoesNotCreateBusinessTables(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	bootstrapper := storage.NewBootstrapper()

	first, err := bootstrapper.Initialize(ctx, storage.StorageConfig{DataDir: root, Mode: storage.ModeTest})
	if err != nil {
		t.Fatalf("first Initialize() error = %v", err)
	}
	second, err := bootstrapper.Initialize(ctx, storage.StorageConfig{DataDir: root, Mode: storage.ModeTest})
	if err != nil {
		t.Fatalf("second Initialize() error = %v", err)
	}

	if second.AppliedMigrations != first.AppliedMigrations || second.PendingMigrations != 0 {
		t.Fatalf("second diagnostics = %+v, want idempotent first result %+v", second, first)
	}
	if second.BusinessTablesCreated != 0 {
		t.Fatalf("BusinessTablesCreated = %d, want 0 in this spec", second.BusinessTablesCreated)
	}
}

func TestStoragePolicyClassifiesConfigBusinessAndSensitiveData(t *testing.T) {
	policy := storage.DefaultPolicy()

	if policy.Classify(storage.DataTheme) != storage.DataClassOrdinaryConfig {
		t.Fatalf("theme class = %q, want ordinary config", policy.Classify(storage.DataTheme))
	}
	if policy.Classify(storage.DataSchemaCache) != storage.DataClassStructuredBusiness {
		t.Fatalf("schema cache class = %q, want structured business", policy.Classify(storage.DataSchemaCache))
	}
	if policy.Classify(storage.DataDatabasePassword) != storage.DataClassSensitiveSecret {
		t.Fatalf("database password class = %q, want sensitive secret", policy.Classify(storage.DataDatabasePassword))
	}
	if !policy.NetworkUploadDisabled {
		t.Fatal("NetworkUploadDisabled = false, want true for local product data")
	}
}

func assertPathEqual(t *testing.T, got string, want string) {
	t.Helper()
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("path = %q, want %q", got, want)
	}
}
