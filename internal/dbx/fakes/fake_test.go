package fakes_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/gerdong/loomidbx/internal/dbx/capability"
	"github.com/gerdong/loomidbx/internal/dbx/core"
	"github.com/gerdong/loomidbx/internal/dbx/dialect"
	"github.com/gerdong/loomidbx/internal/dbx/fakes"
	"github.com/gerdong/loomidbx/internal/dbx/introspect"
	"github.com/gerdong/loomidbx/internal/dbx/schema"
	"github.com/gerdong/loomidbx/internal/dbx/typex"
)

func TestFakeAdapterCompositionThroughRegistry(t *testing.T) {
	ctx := context.Background()
	registry := core.NewRegistry()
	adapter := fakes.NewAdapter(core.DBTypePostgres)
	adapter.Caps = capability.PostgreSQLExample()

	if err := registry.Register(adapter); err != nil {
		t.Fatalf("Register(fake) error = %v", err)
	}
	resolved, err := registry.Get(core.DBTypePostgres)
	if err != nil {
		t.Fatalf("Get(fake) error = %v", err)
	}

	if !resolved.Capabilities().SupportsReturning {
		t.Fatal("fake adapter capabilities must be available through registry")
	}
	result := resolved.TestConnection(ctx, core.ConnectionConfig{Type: core.DBTypePostgres, Host: "localhost", Password: "test-only"})
	if !result.OK {
		t.Fatalf("TestConnection() = %#v, want OK", result)
	}

	database, err := resolved.Introspector().Introspect(ctx, nil, introspect.Options{Schema: "public"})
	if err != nil {
		t.Fatalf("Introspect() error = %v", err)
	}
	again, err := resolved.Introspector().Introspect(ctx, nil, introspect.Options{Schema: "public"})
	if err != nil {
		t.Fatalf("second Introspect() error = %v", err)
	}
	if !reflect.DeepEqual(database, again) {
		t.Fatal("fake introspector must return deterministic schema")
	}

	logical := resolved.TypeMapper().ToLogical(typex.NativeType{Name: "integer"}, typex.MappingOptions{})
	if logical.Kind != schema.KindInteger {
		t.Fatalf("ToLogical(integer).Kind = %q, want integer", logical.Kind)
	}

	statements, err := resolved.Dialect().BuildInsert(dialect.InsertRequest{
		Table:   "users",
		Columns: []schema.Column{{Name: "id"}},
		Rows:    []map[string]any{{"id": 1}},
	})
	if err != nil {
		t.Fatalf("BuildInsert() error = %v", err)
	}
	if statements[0].SQL == "" || len(statements[0].Args) != 1 {
		t.Fatalf("statement = %#v, want SQL and one arg", statements[0])
	}

	if adapter.TestConnectionCalls != 1 || adapter.IntrospectorCalls != 2 || adapter.TypeMapperCalls != 1 || adapter.DialectCalls != 1 {
		t.Fatalf("adapter call counts = test:%d introspector:%d mapper:%d dialect:%d", adapter.TestConnectionCalls, adapter.IntrospectorCalls, adapter.TypeMapperCalls, adapter.DialectCalls)
	}
}
