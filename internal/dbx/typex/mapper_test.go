package typex_test

import (
	"testing"

	"github.com/gerdong/loomidbx/internal/dbx/fakes"
	"github.com/gerdong/loomidbx/internal/dbx/schema"
	"github.com/gerdong/loomidbx/internal/dbx/typex"
)

func TestFakeMapperMapsKnownNativeType(t *testing.T) {
	mapper := fakes.NewMapper()
	length := int64(255)

	logical := mapper.ToLogical(typex.NativeType{Full: "varchar(255)", Length: &length}, typex.MappingOptions{PreserveNativeType: true})

	if logical.Kind != schema.KindString {
		t.Fatalf("logical.Kind = %q, want %q", logical.Kind, schema.KindString)
	}
	if logical.NativeType != "varchar(255)" {
		t.Fatalf("logical.NativeType = %q, want native type preserved", logical.NativeType)
	}
	if mapper.Calls != 1 {
		t.Fatalf("mapper.Calls = %d, want 1", mapper.Calls)
	}
}

func TestFakeMapperUsesConfiguredFallbackForUnknownNativeType(t *testing.T) {
	mapper := fakes.NewMapper()

	logical := mapper.ToLogical(typex.NativeType{Name: "geometry"}, typex.MappingOptions{Fallback: schema.KindString})

	if logical.Kind != schema.KindString || logical.NativeType != "geometry" {
		t.Fatalf("logical = %#v, want fallback string preserving native type", logical)
	}
}
