package schema

import (
	"reflect"
	"testing"
	"time"
)

func TestTableFieldConstraintScaffoldExportsStableShapes(t *testing.T) {
	assertJSONTags(t, reflect.TypeOf(DbTable{}), map[string]string{
		"ID":          "id",
		"SchemaID":    "schemaId",
		"TableName":   "tableName",
		"Comment":     "comment",
		"DDLSnapshot": "ddlSnapshot",
		"ScannedAt":   "scannedAt",
		"CreatedAt":   "createdAt",
		"UpdatedAt":   "updatedAt",
	})
	assertStructJSONFieldSet(t, reflect.TypeOf(DbTable{}), []string{"id", "schemaId", "tableName", "comment", "ddlSnapshot", "scannedAt", "createdAt", "updatedAt"})

	assertJSONTags(t, reflect.TypeOf(DbColumn{}), map[string]string{
		"ID":              "id",
		"TableID":         "tableId",
		"OrdinalPosition": "ordinalPosition",
		"ColumnName":      "columnName",
		"NativeType":      "nativeType",
		"LogicalType":     "logicalType",
		"Nullable":        "nullable",
		"DefaultValue":    "defaultValue",
		"IsPrimaryKey":    "isPrimaryKey",
		"Comment":         "comment",
		"CreatedAt":       "createdAt",
		"UpdatedAt":       "updatedAt",
	})
	assertStructJSONFieldSet(t, reflect.TypeOf(DbColumn{}), []string{"id", "tableId", "ordinalPosition", "columnName", "nativeType", "logicalType", "nullable", "defaultValue", "isPrimaryKey", "comment", "createdAt", "updatedAt"})

	assertJSONTags(t, reflect.TypeOf(TableConstraint{}), map[string]string{
		"ID":             "id",
		"TableID":        "tableId",
		"ConstraintName": "constraintName",
		"ConstraintType": "constraintType",
		"ColumnIDs":      "columnIds",
		"CreatedAt":      "createdAt",
	})
	assertStructJSONFieldSet(t, reflect.TypeOf(TableConstraint{}), []string{"id", "tableId", "constraintName", "constraintType", "columnIds", "createdAt"})

	assertJSONTags(t, reflect.TypeOf(ColumnLogicalType{}), map[string]string{
		"Kind":       "kind",
		"Length":     "length",
		"Precision":  "precision",
		"Scale":      "scale",
		"BitWidth":   "bitWidth",
		"Timezone":   "timezone",
		"Element":    "element",
		"EnumValues": "enumValues",
		"NativeType": "nativeType",
	})
	assertStructJSONFieldSet(t, reflect.TypeOf(ColumnLogicalType{}), []string{"kind", "length", "precision", "scale", "bitWidth", "timezone", "element", "enumValues", "nativeType"})
}

func TestTableFieldConstraintScaffoldDeclaresStableEnums(t *testing.T) {
	tests := map[string]string{
		"TableConstraintTypePrimary": string(TableConstraintTypePrimary),
		"TableConstraintTypeUnique":  string(TableConstraintTypeUnique),
		"ColumnLogicalKindUnknown":   string(ColumnLogicalKindUnknown),
		"ColumnLogicalKindString":    string(ColumnLogicalKindString),
		"ColumnLogicalKindText":      string(ColumnLogicalKindText),
		"ColumnLogicalKindInteger":   string(ColumnLogicalKindInteger),
		"ColumnLogicalKindDecimal":   string(ColumnLogicalKindDecimal),
		"ColumnLogicalKindFloat":     string(ColumnLogicalKindFloat),
		"ColumnLogicalKindBoolean":   string(ColumnLogicalKindBoolean),
		"ColumnLogicalKindDate":      string(ColumnLogicalKindDate),
		"ColumnLogicalKindTime":      string(ColumnLogicalKindTime),
		"ColumnLogicalKindDateTime":  string(ColumnLogicalKindDateTime),
		"ColumnLogicalKindBinary":    string(ColumnLogicalKindBinary),
		"ColumnLogicalKindJSON":      string(ColumnLogicalKindJSON),
		"ColumnLogicalKindUUID":      string(ColumnLogicalKindUUID),
		"ColumnLogicalKindArray":     string(ColumnLogicalKindArray),
		"ColumnLogicalKindEnum":      string(ColumnLogicalKindEnum),
	}

	expected := map[string]string{
		"TableConstraintTypePrimary": "PRIMARY",
		"TableConstraintTypeUnique":  "UNIQUE",
		"ColumnLogicalKindUnknown":   "unknown",
		"ColumnLogicalKindString":    "string",
		"ColumnLogicalKindText":      "text",
		"ColumnLogicalKindInteger":   "integer",
		"ColumnLogicalKindDecimal":   "decimal",
		"ColumnLogicalKindFloat":     "float",
		"ColumnLogicalKindBoolean":   "boolean",
		"ColumnLogicalKindDate":      "date",
		"ColumnLogicalKindTime":      "time",
		"ColumnLogicalKindDateTime":  "datetime",
		"ColumnLogicalKindBinary":    "binary",
		"ColumnLogicalKindJSON":      "json",
		"ColumnLogicalKindUUID":      "uuid",
		"ColumnLogicalKindArray":     "array",
		"ColumnLogicalKindEnum":      "enum",
	}

	for name, got := range tests {
		if want := expected[name]; got != want {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestTableFieldConstraintScaffoldSerializesStableJSONContracts(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	defaultValue := "nextval('users_id_seq')"
	length := int64(255)

	column := DbColumn{
		ID:              20,
		TableID:         10,
		OrdinalPosition: 1,
		ColumnName:      "email",
		NativeType:      "varchar(255)",
		LogicalType:     ColumnLogicalType{Kind: ColumnLogicalKindString, Length: &length, NativeType: "varchar(255)"},
		Nullable:        false,
		DefaultValue:    &defaultValue,
		IsPrimaryKey:    true,
		Comment:         "business email",
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
	}
	assertJSONRoundTrip(t, "DbColumn", column)

	table := DbTable{
		ID:          10,
		SchemaID:    5,
		TableName:   "users",
		Comment:     "application users",
		DDLSnapshot: "CREATE TABLE users (...) ",
		ScannedAt:   &createdAt,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
	assertJSONRoundTrip(t, "DbTable", table)

	constraint := TableConstraint{
		ID:             30,
		TableID:        10,
		ConstraintName: "users_pkey",
		ConstraintType: TableConstraintTypePrimary,
		ColumnIDs:      []int64{20},
		CreatedAt:      createdAt,
	}
	assertJSONRoundTrip(t, "TableConstraint", constraint)
}
