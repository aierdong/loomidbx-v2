package core_test

import (
	"errors"
	"testing"

	"github.com/gerdong/loomidbx/internal/dbx/core"
	"github.com/gerdong/loomidbx/internal/dbx/fakes"
)

func TestRegistryRegistersGetsAndListsAdapters(t *testing.T) {
	registry := core.NewRegistry()
	postgres := fakes.NewAdapter(core.DBTypePostgres)
	mysql := fakes.NewAdapter(core.DBTypeMySQL)

	if err := registry.Register(postgres); err != nil {
		t.Fatalf("Register(postgres) error = %v", err)
	}
	if err := registry.Register(mysql); err != nil {
		t.Fatalf("Register(mysql) error = %v", err)
	}

	adapter, err := registry.Get(core.DBTypePostgres)
	if err != nil {
		t.Fatalf("Get(postgres) error = %v", err)
	}
	if adapter.Type() != core.DBTypePostgres {
		t.Fatalf("adapter.Type() = %q, want %q", adapter.Type(), core.DBTypePostgres)
	}

	infos := registry.List()
	if len(infos) != 2 {
		t.Fatalf("len(List()) = %d, want 2", len(infos))
	}
	if infos[0].Type != core.DBTypeMySQL || infos[1].Type != core.DBTypePostgres {
		t.Fatalf("List() order = %#v, want mysql then postgres", infos)
	}
}

func TestRegistryReturnsTypedErrors(t *testing.T) {
	registry := core.NewRegistry()

	if err := registry.Register(nil); !errors.Is(err, &core.DBXError{Kind: core.ErrorInvalidAdapter}) {
		t.Fatalf("Register(nil) error = %v, want invalid adapter", err)
	}

	empty := fakes.NewAdapter("")
	if err := registry.Register(empty); !errors.Is(err, &core.DBXError{Kind: core.ErrorInvalidAdapter}) {
		t.Fatalf("Register(empty type) error = %v, want invalid adapter", err)
	}

	mysql := fakes.NewAdapter(core.DBTypeMySQL)
	if err := registry.Register(mysql); err != nil {
		t.Fatalf("Register(mysql) error = %v", err)
	}
	if err := registry.Register(mysql); !errors.Is(err, &core.DBXError{Kind: core.ErrorDuplicateAdapter}) {
		t.Fatalf("Register(duplicate) error = %v, want duplicate adapter", err)
	}

	adapter, err := registry.Get("sqlite")
	if adapter != nil {
		t.Fatalf("Get(sqlite) adapter = %#v, want nil", adapter)
	}
	if !errors.Is(err, &core.DBXError{Kind: core.ErrorUnsupportedDatabase}) {
		t.Fatalf("Get(sqlite) error = %v, want unsupported database", err)
	}
}

func TestTypedErrorsDoNotExposeSensitiveConnectionData(t *testing.T) {
	err := core.NewError(core.ErrorInvalidConnectionConfig, core.DBTypeMySQL, "test connection", "database inventory", errors.New("missing host"))
	message := err.Error()
	for _, forbidden := range []string{"password", "secret", "token", "dsn=", "user SQL", "args"} {
		if containsFold(message, forbidden) {
			t.Fatalf("error message %q contains forbidden text %q", message, forbidden)
		}
	}
}

func containsFold(value string, needle string) bool {
	return len(needle) == 0 || stringsContainsFold(value, needle)
}

func stringsContainsFold(value string, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if equalFold(value[i:i+len(needle)], needle) {
			return true
		}
	}
	return false
}

func equalFold(left string, right string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		l := left[i]
		r := right[i]
		if 'A' <= l && l <= 'Z' {
			l += 'a' - 'A'
		}
		if 'A' <= r && r <= 'Z' {
			r += 'a' - 'A'
		}
		if l != r {
			return false
		}
	}
	return true
}
