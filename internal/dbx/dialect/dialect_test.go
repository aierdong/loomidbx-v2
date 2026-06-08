package dialect_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gerdong/loomidbx/internal/dbx/core"
	"github.com/gerdong/loomidbx/internal/dbx/dialect"
	"github.com/gerdong/loomidbx/internal/dbx/fakes"
	"github.com/gerdong/loomidbx/internal/dbx/schema"
)

func TestFakeDialectQuotesIdentifierAndPlaceholder(t *testing.T) {
	fake := fakes.NewDialect()
	fake.PlaceholderPrefix = "$"

	if got := fake.QuoteIdentifier("user\"name"); got != "\"user\"\"name\"" {
		t.Fatalf("QuoteIdentifier() = %q, want escaped double quotes", got)
	}
	if got := fake.Placeholder(3); got != "$3" {
		t.Fatalf("Placeholder(3) = %q, want $3", got)
	}
}

func TestFakeDialectBuildInsertReturnsStatementAndArgs(t *testing.T) {
	fake := fakes.NewDialect()
	req := dialect.InsertRequest{
		Schema:  "public",
		Table:   "users",
		Columns: []schema.Column{{Name: "id"}, {Name: "email"}},
		Rows: []map[string]any{
			{"id": 1, "email": "a@example.test"},
			{"id": 2, "email": "b@example.test"},
		},
	}

	statements, err := fake.BuildInsert(req)
	if err != nil {
		t.Fatalf("BuildInsert() error = %v", err)
	}
	if len(statements) != 1 {
		t.Fatalf("len(statements) = %d, want 1", len(statements))
	}
	wantSQL := "INSERT INTO \"public\".\"users\" (\"id\", \"email\") VALUES (?, ?), (?, ?)"
	if statements[0].SQL != wantSQL {
		t.Fatalf("SQL = %q, want %q", statements[0].SQL, wantSQL)
	}
	if !reflect.DeepEqual(statements[0].Args, []any{1, "a@example.test", 2, "b@example.test"}) {
		t.Fatalf("Args = %#v", statements[0].Args)
	}
}

func TestFakeDialectReturnsTypedUnsupportedOperation(t *testing.T) {
	fake := fakes.NewDialect()
	_, err := fake.BuildInsert(dialect.InsertRequest{Table: "users"})
	if !errors.Is(err, &core.DBXError{Kind: core.ErrorUnsupportedDialectOperation}) {
		t.Fatalf("BuildInsert(invalid) error = %v, want unsupported dialect operation", err)
	}
}
