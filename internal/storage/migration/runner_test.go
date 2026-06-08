package migration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gerdong/loomidbx/internal/storage/migration"
)

func TestRegistrySortsMigrationsByVersion(t *testing.T) {
	registry := migration.NewRegistry([]migration.Migration{
		{Version: "000002", Name: "second", Up: "SELECT 2"},
		{Version: "000001", Name: "first", Up: "SELECT 1"},
	})

	migrations, err := registry.List()

	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if migrations[0].Version != "000001" || migrations[1].Version != "000002" {
		t.Fatalf("versions = %q, %q, want sorted", migrations[0].Version, migrations[1].Version)
	}
}

func TestRegistryRejectsInvalidMigrationName(t *testing.T) {
	registry := migration.NewRegistry([]migration.Migration{
		{Version: "1", Name: "bad", Up: "SELECT 1"},
	})

	_, err := registry.List()

	var migrationErr *migration.MigrationError
	if !errors.As(err, &migrationErr) {
		t.Fatalf("List() error = %T %[1]v, want MigrationError", err)
	}
	if migrationErr.Code != "STORAGE_MIGRATION_FAILED" {
		t.Fatalf("Code = %q, want STORAGE_MIGRATION_FAILED", migrationErr.Code)
	}
}

func TestRunnerAppliesOnlyPendingMigrationsAndRecordsSuccess(t *testing.T) {
	ctx := context.Background()
	conn := migration.NewMemoryConnection()
	registry := migration.NewRegistry([]migration.Migration{
		{Version: "000001", Name: "first", Up: "CREATE TABLE first_table (id TEXT)"},
		{Version: "000002", Name: "second", Up: "CREATE TABLE second_table (id TEXT)"},
	})
	runner := migration.NewRunner(registry)

	first, err := runner.Apply(ctx, conn)
	if err != nil {
		t.Fatalf("first Apply() error = %v, want nil", err)
	}
	second, err := runner.Apply(ctx, conn)
	if err != nil {
		t.Fatalf("second Apply() error = %v, want nil", err)
	}

	if first.AppliedCount != 2 || first.PendingCount != 0 || first.CurrentVersion != "000002" {
		t.Fatalf("first result = %+v, want two applied", first)
	}
	if second.AppliedCount != 0 || second.PendingCount != 0 || second.CurrentVersion != "000002" {
		t.Fatalf("second result = %+v, want no duplicate execution", second)
	}
	if conn.ExecCount("CREATE TABLE first_table (id TEXT)") != 1 {
		t.Fatalf("first migration exec count = %d, want 1", conn.ExecCount("CREATE TABLE first_table (id TEXT)"))
	}
}

func TestRunnerDoesNotRecordFailedMigrationAndAllowsRetry(t *testing.T) {
	ctx := context.Background()
	conn := migration.NewMemoryConnection()
	failing := migration.NewRegistry([]migration.Migration{
		{Version: "000001", Name: "first", Up: "CREATE TABLE first_table (id TEXT)"},
		{Version: "000002", Name: "second", Up: "FAIL with password=secret-token"},
	})
	runner := migration.NewRunner(failing)

	_, err := runner.Apply(ctx, conn)

	var migrationErr *migration.MigrationError
	if !errors.As(err, &migrationErr) {
		t.Fatalf("Apply() error = %T %[1]v, want MigrationError", err)
	}
	if migrationErr.Code != "STORAGE_MIGRATION_FAILED" {
		t.Fatalf("Code = %q, want STORAGE_MIGRATION_FAILED", migrationErr.Code)
	}
	if migrationErr.MigrationVersion != "000002" {
		t.Fatalf("MigrationVersion = %q, want 000002", migrationErr.MigrationVersion)
	}
	if conn.HasApplied("000002") {
		t.Fatal("failed migration was recorded as applied")
	}
	if migrationErr.Error() == "" || containsSensitive(migrationErr.Error()) {
		t.Fatalf("Error() = %q, want sanitized message", migrationErr.Error())
	}

	fixed := migration.NewRunner(migration.NewRegistry([]migration.Migration{
		{Version: "000001", Name: "first", Up: "CREATE TABLE first_table (id TEXT)"},
		{Version: "000002", Name: "second", Up: "CREATE TABLE second_table (id TEXT)"},
	}))
	result, retryErr := fixed.Apply(ctx, conn)
	if retryErr != nil {
		t.Fatalf("retry Apply() error = %v, want nil", retryErr)
	}
	if result.AppliedCount != 1 || result.CurrentVersion != "000002" || !conn.HasApplied("000002") {
		t.Fatalf("retry result = %+v, want second migration applied", result)
	}
}

func containsSensitive(value string) bool {
	return stringsContains(value, "secret-token") || stringsContains(value, "password=")
}

func stringsContains(value string, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
