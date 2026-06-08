package schema_test

import (
	"testing"

	"github.com/gerdong/loomidbx/internal/dbx/schema"
)

func TestCanonicalSchemaPreservesColumnAndRawMetadata(t *testing.T) {
	length := int64(255)
	database := schema.Database{
		Name: "inventory",
		Raw:  schema.RawMetadata{"server_version": "test-only"},
		Namespaces: []schema.Namespace{{
			Name: "public",
			Tables: []schema.Table{{
				Name: "items",
				Columns: []schema.Column{{
					Name:       "sku",
					NativeType: "varchar(255)",
					LogicalType: schema.LogicalType{
						Kind:       schema.KindString,
						Length:     &length,
						NativeType: "varchar(255)",
					},
					Nullable: false,
					Ordinal:  1,
					Comment:  "stable identifier",
					Raw:      schema.RawMetadata{"collation": "utf8mb4_bin"},
				}},
				PrimaryKey:  &schema.PrimaryKey{Name: "items_pkey", Columns: []string{"sku"}},
				ForeignKeys: []schema.ForeignKey{{Name: "items_parent_fk", Columns: []string{"sku"}, ReferencedTable: "parents", ReferencedColumns: []string{"sku"}}},
			}},
		}},
	}

	column := database.Namespaces[0].Tables[0].Columns[0]
	if column.NativeType != "varchar(255)" || column.LogicalType.Kind != schema.KindString {
		t.Fatalf("column type = %#v, want native varchar and logical string", column)
	}
	if column.Raw["collation"] != "utf8mb4_bin" {
		t.Fatalf("column.Raw = %#v, want collation metadata", column.Raw)
	}
	if database.Namespaces[0].Tables[0].PrimaryKey.Columns[0] != "sku" {
		t.Fatal("primary key metadata must be preserved")
	}
	if len(database.Namespaces[0].Tables[0].ForeignKeys) != 1 {
		t.Fatal("foreign key metadata must be preserved")
	}
}

func TestUnknownLogicalTypePreservesNativeType(t *testing.T) {
	logical := schema.UnknownLogicalType("geometry(point)")
	if logical.Kind != schema.KindUnknown || logical.NativeType != "geometry(point)" {
		t.Fatalf("UnknownLogicalType = %#v, want unknown with native type", logical)
	}
}
