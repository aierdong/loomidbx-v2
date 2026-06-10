package schema

import (
	"encoding/json"
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

	for _, known := range []TableConstraintType{TableConstraintTypePrimary, TableConstraintTypeUnique} {
		if !known.IsKnown() {
			t.Fatalf("%q should be recognized as known", known)
		}
		if known.IsUnknown() {
			t.Fatalf("%q should not be recognized as unknown", known)
		}
	}

	unknownConstraint := TableConstraintType("FOREIGN_KEY")
	if unknownConstraint.IsKnown() {
		t.Fatalf("%q should not be recognized as known", unknownConstraint)
	}
	if !unknownConstraint.IsUnknown() {
		t.Fatalf("%q should be recognized as unknown", unknownConstraint)
	}

	for _, known := range []ColumnLogicalKind{
		ColumnLogicalKindUnknown,
		ColumnLogicalKindString,
		ColumnLogicalKindText,
		ColumnLogicalKindInteger,
		ColumnLogicalKindDecimal,
		ColumnLogicalKindFloat,
		ColumnLogicalKindBoolean,
		ColumnLogicalKindDate,
		ColumnLogicalKindTime,
		ColumnLogicalKindDateTime,
		ColumnLogicalKindBinary,
		ColumnLogicalKindJSON,
		ColumnLogicalKindUUID,
		ColumnLogicalKindArray,
		ColumnLogicalKindEnum,
	} {
		if !known.IsKnown() {
			t.Fatalf("%q should be recognized as known", known)
		}
		if known.IsUnknown() {
			t.Fatalf("%q should not be recognized as unknown", known)
		}
	}

	unknownKind := ColumnLogicalKind("geometry")
	if unknownKind.IsKnown() {
		t.Fatalf("%q should not be recognized as known", unknownKind)
	}
	if !unknownKind.IsUnknown() {
		t.Fatalf("%q should be recognized as unknown", unknownKind)
	}
}

func TestDomainEnumJSONPreservesUnknownStringsAndRejectsNonStrings(t *testing.T) {
	var constraintType TableConstraintType
	if err := json.Unmarshal([]byte(`"CHECK"`), &constraintType); err != nil {
		t.Fatalf("Unmarshal unknown TableConstraintType returned error: %v", err)
	}
	if constraintType != TableConstraintType("CHECK") || !constraintType.IsUnknown() {
		t.Fatalf("unknown TableConstraintType = %q, IsUnknown=%v; want preserved unknown", constraintType, constraintType.IsUnknown())
	}
	encodedConstraintType, err := json.Marshal(constraintType)
	if err != nil {
		t.Fatalf("Marshal unknown TableConstraintType returned error: %v", err)
	}
	if string(encodedConstraintType) != `"CHECK"` {
		t.Fatalf("unknown TableConstraintType JSON = %s, want %q", encodedConstraintType, `"CHECK"`)
	}

	var logicalKind ColumnLogicalKind
	if err := json.Unmarshal([]byte(`"geometry"`), &logicalKind); err != nil {
		t.Fatalf("Unmarshal unknown ColumnLogicalKind returned error: %v", err)
	}
	if logicalKind != ColumnLogicalKind("geometry") || !logicalKind.IsUnknown() {
		t.Fatalf("unknown ColumnLogicalKind = %q, IsUnknown=%v; want preserved unknown", logicalKind, logicalKind.IsUnknown())
	}
	encodedLogicalKind, err := json.Marshal(logicalKind)
	if err != nil {
		t.Fatalf("Marshal unknown ColumnLogicalKind returned error: %v", err)
	}
	if string(encodedLogicalKind) != `"geometry"` {
		t.Fatalf("unknown ColumnLogicalKind JSON = %s, want %q", encodedLogicalKind, `"geometry"`)
	}

	for name, decode := range map[string]func() error{
		"constraint type number": func() error {
			var decoded TableConstraintType
			return json.Unmarshal([]byte(`1`), &decoded)
		},
		"logical kind object": func() error {
			var decoded ColumnLogicalKind
			return json.Unmarshal([]byte(`{"kind":"string"}`), &decoded)
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := decode(); err == nil {
				t.Fatalf("enum unmarshal should reject non-string JSON for %s", name)
			}
		})
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
