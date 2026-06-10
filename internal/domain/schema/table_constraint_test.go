package schema

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
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

func TestTableValidationReusesUpstreamStructuredIssuesAndReturnsMultipleProblems(t *testing.T) {
	var _ []SchemaValidationIssue = ValidateTable(DbTable{}, SchemaValidationModeDraft)
	var _ []SchemaValidationIssue = ValidateColumn(DbColumn{}, SchemaValidationModeDraft)
	var _ []SchemaValidationIssue = ValidateConstraint(TableConstraint{}, SchemaValidationModeDraft)
	var _ []SchemaValidationIssue = ValidateLogicalType(ColumnLogicalType{})

	issues := ValidateConstraint(TableConstraint{
		TableID:        10,
		ConstraintName: "users_email_key",
		ConstraintType: TableConstraintType("CHECK"),
		ColumnIDs:      []int64{20},
	}, SchemaValidationMode("runtime"))

	assertIssuePaths(t, issues, []string{"mode", "constraintType"})
	assertIssueCodes(t, issues, map[string]SchemaIssueCode{
		"mode":           SchemaIssueCodeValidationFailed,
		"constraintType": SchemaIssueCodeValidationFailed,
	})
	assertAllIssuesSafeErrors(t, issues)
	for _, issue := range issues {
		if contractIssues := ValidateIssue(issue); len(contractIssues) != 0 {
			t.Fatalf("table validation issue does not satisfy upstream issue contract: %#v produced %#v", issue, contractIssues)
		}
	}
}

func TestValidateTableAppliesRequiredParentRangeAndAuditRules(t *testing.T) {
	if issues := ValidateTable(DbTable{SchemaID: 5, TableName: "users"}, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateTable(valid draft) = %#v, want no issues", issues)
	}

	createdAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(-time.Second)
	zeroScannedAt := time.Time{}
	issues := ValidateTable(DbTable{
		ID:        -1,
		SchemaID:  0,
		TableName: " /bad",
		ScannedAt: &zeroScannedAt,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, SchemaValidationModeDraft)

	assertIssuePaths(t, issues, []string{"id", "schemaId", "tableName", "scannedAt", "updatedAt"})
	assertIssueCodes(t, issues, map[string]SchemaIssueCode{
		"id":        SchemaIssueCodeInvalidID,
		"schemaId":  SchemaIssueCodeInvalidID,
		"tableName": SchemaIssueCodeInvalidName,
		"scannedAt": SchemaIssueCodeInvalidTime,
		"updatedAt": SchemaIssueCodeInvalidTime,
	})
	assertAllIssuesSafeErrors(t, issues)

	invalidModeIssues := ValidateTable(DbTable{SchemaID: 0, TableName: ""}, SchemaValidationMode("runtime"))
	assertIssuePaths(t, invalidModeIssues, []string{"mode", "schemaId", "tableName"})

	persistedIssues := ValidateTable(DbTable{SchemaID: 5, TableName: "users"}, SchemaValidationModePersisted)
	assertIssuePaths(t, persistedIssues, []string{"id", "createdAt", "updatedAt"})
}

func TestValidateColumnAppliesRequiredParentEnumRangeAndAuditRules(t *testing.T) {
	if issues := ValidateColumn(DbColumn{
		TableID:         10,
		OrdinalPosition: 1,
		ColumnName:      "email",
		NativeType:      "varchar(255)",
		LogicalType:     ColumnLogicalType{Kind: ColumnLogicalKindString},
	}, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateColumn(valid draft) = %#v, want no issues", issues)
	}

	issues := ValidateColumn(DbColumn{
		ID:              -1,
		TableID:         0,
		OrdinalPosition: 0,
		ColumnName:      "bad/name",
		NativeType:      "   ",
		LogicalType:     ColumnLogicalType{Kind: ColumnLogicalKind("geometry")},
	}, SchemaValidationModeDraft)

	assertIssuePaths(t, issues, []string{"id", "tableId", "ordinalPosition", "columnName", "nativeType", "logicalType.kind"})
	assertIssueCodes(t, issues, map[string]SchemaIssueCode{
		"id":               SchemaIssueCodeInvalidID,
		"tableId":          SchemaIssueCodeInvalidID,
		"ordinalPosition":  SchemaIssueCodeValidationFailed,
		"columnName":       SchemaIssueCodeInvalidName,
		"nativeType":       SchemaIssueCodeRequired,
		"logicalType.kind": SchemaIssueCodeValidationFailed,
	})
	assertAllIssuesSafeErrors(t, issues)

	invalidModeIssues := ValidateColumn(DbColumn{
		TableID:         0,
		OrdinalPosition: 0,
		ColumnName:      "",
		NativeType:      "",
		LogicalType:     ColumnLogicalType{Kind: ColumnLogicalKind("geometry")},
	}, SchemaValidationMode("runtime"))
	assertIssuePaths(t, invalidModeIssues, []string{"mode", "tableId", "ordinalPosition", "columnName", "nativeType", "logicalType.kind"})

	persistedIssues := ValidateColumn(DbColumn{
		TableID:         10,
		OrdinalPosition: 1,
		ColumnName:      "email",
		NativeType:      "varchar(255)",
		LogicalType:     ColumnLogicalType{Kind: ColumnLogicalKindString},
	}, SchemaValidationModePersisted)
	assertIssuePaths(t, persistedIssues, []string{"id", "createdAt", "updatedAt"})
}

func TestValidateConstraintAppliesRequiredParentEnumRangeAndInternalUniquenessRules(t *testing.T) {
	if issues := ValidateConstraint(TableConstraint{
		TableID:        10,
		ConstraintName: "users_email_key",
		ConstraintType: TableConstraintTypeUnique,
		ColumnIDs:      []int64{20},
	}, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateConstraint(valid draft) = %#v, want no issues", issues)
	}

	issues := ValidateConstraint(TableConstraint{
		ID:             -1,
		TableID:        0,
		ConstraintName: "",
		ConstraintType: TableConstraintType("CHECK"),
		ColumnIDs:      []int64{0, 20, 20},
	}, SchemaValidationModeDraft)

	assertIssuePaths(t, issues, []string{"id", "tableId", "constraintName", "constraintType", "columnIds", "columnIds"})
	assertAllIssuesSafeErrors(t, issues)

	emptyColumnIssues := ValidateConstraint(TableConstraint{
		TableID:        10,
		ConstraintName: "users_email_key",
		ConstraintType: TableConstraintTypeUnique,
	}, SchemaValidationModeDraft)
	assertIssuePaths(t, emptyColumnIssues, []string{"columnIds"})
	assertIssueCodes(t, emptyColumnIssues, map[string]SchemaIssueCode{"columnIds": SchemaIssueCodeRequired})

	persistedIssues := ValidateConstraint(TableConstraint{
		TableID:        10,
		ConstraintName: "users_email_key",
		ConstraintType: TableConstraintTypeUnique,
		ColumnIDs:      []int64{20},
	}, SchemaValidationModePersisted)
	assertIssuePaths(t, persistedIssues, []string{"id", "createdAt"})
}

func TestValidateLogicalTypeAppliesEnumRangeAndSingleObjectUniquenessRules(t *testing.T) {
	length := int64(1)
	precision := 2
	scale := 1
	bitWidth := 32
	if issues := ValidateLogicalType(ColumnLogicalType{
		Kind:       ColumnLogicalKindDecimal,
		Length:     &length,
		Precision:  &precision,
		Scale:      &scale,
		BitWidth:   &bitWidth,
		NativeType: "numeric(2,1)",
	}); len(issues) != 0 {
		t.Fatalf("ValidateLogicalType(valid decimal) = %#v, want no issues", issues)
	}

	zeroLength := int64(0)
	zeroPrecision := 0
	largeScale := 3
	zeroBitWidth := 0
	issues := ValidateLogicalType(ColumnLogicalType{
		Kind:      ColumnLogicalKind("geometry"),
		Length:    &zeroLength,
		Precision: &zeroPrecision,
		Scale:     &largeScale,
		BitWidth:  &zeroBitWidth,
	})
	assertIssuePaths(t, issues, []string{"kind", "length", "precision", "bitWidth"})
	assertAllIssuesSafeErrors(t, issues)

	precisionTwo := 2
	scaleThree := 3
	scaleIssues := ValidateLogicalType(ColumnLogicalType{Kind: ColumnLogicalKindDecimal, Precision: &precisionTwo, Scale: &scaleThree})
	assertIssuePaths(t, scaleIssues, []string{"scale"})

	unknownIssues := ValidateLogicalType(ColumnLogicalType{Kind: ColumnLogicalKindUnknown, NativeType: "   "})
	assertIssuePaths(t, unknownIssues, []string{"nativeType"})

	emptyEnumIssues := ValidateLogicalType(ColumnLogicalType{Kind: ColumnLogicalKindEnum})
	assertIssuePaths(t, emptyEnumIssues, []string{"enumValues"})

	duplicateEnumIssues := ValidateLogicalType(ColumnLogicalType{Kind: ColumnLogicalKindEnum, EnumValues: []string{"small", " ", "small"}})
	assertIssuePaths(t, duplicateEnumIssues, []string{"enumValues", "enumValues"})

	arrayIssues := ValidateLogicalType(ColumnLogicalType{Kind: ColumnLogicalKindArray})
	assertIssuePaths(t, arrayIssues, []string{"element"})
}

func TestTableFieldConstraintMissingOptionalFieldsDecodeToCompatibleDefaults(t *testing.T) {
	var table DbTable
	if err := json.Unmarshal([]byte(`{"id":1,"schemaId":2,"tableName":"users"}`), &table); err != nil {
		t.Fatalf("Unmarshal minimal DbTable returned error: %v", err)
	}
	if table.ID != 1 || table.SchemaID != 2 || table.TableName != "users" {
		t.Fatalf("decoded minimal table = %#v, want stable required fields", table)
	}
	if table.Comment != "" || table.DDLSnapshot != "" || table.ScannedAt != nil || !table.CreatedAt.IsZero() || !table.UpdatedAt.IsZero() {
		t.Fatalf("missing optional table fields should decode to compatible zero values: %#v", table)
	}

	var column DbColumn
	if err := json.Unmarshal([]byte(`{"id":3,"tableId":1,"ordinalPosition":1,"columnName":"email","nativeType":"varchar(255)","logicalType":{"kind":"string"}}`), &column); err != nil {
		t.Fatalf("Unmarshal minimal DbColumn returned error: %v", err)
	}
	if column.ID != 3 || column.TableID != 1 || column.OrdinalPosition != 1 || column.ColumnName != "email" || column.NativeType != "varchar(255)" {
		t.Fatalf("decoded minimal column = %#v, want stable required fields", column)
	}
	if column.Nullable || column.IsPrimaryKey || column.DefaultValue != nil || column.Comment != "" || !column.CreatedAt.IsZero() || !column.UpdatedAt.IsZero() {
		t.Fatalf("missing optional column fields should decode compatibly: %#v", column)
	}
	if column.LogicalType.Kind != ColumnLogicalKindString || column.LogicalType.Length != nil || column.LogicalType.EnumValues != nil {
		t.Fatalf("decoded minimal logical type = %#v, want string kind with nil optional metadata", column.LogicalType)
	}

	var logical ColumnLogicalType
	if err := json.Unmarshal([]byte(`{"kind":"unknown","nativeType":"citext"}`), &logical); err != nil {
		t.Fatalf("Unmarshal minimal ColumnLogicalType returned error: %v", err)
	}
	if logical.Kind != ColumnLogicalKindUnknown || logical.NativeType != "citext" || logical.Timezone {
		t.Fatalf("decoded minimal logical type = %#v, want unknown citext without timezone", logical)
	}
	if logical.Length != nil || logical.Precision != nil || logical.Scale != nil || logical.BitWidth != nil || logical.Element != nil || logical.EnumValues != nil {
		t.Fatalf("missing optional logical metadata should decode to nil values: %#v", logical)
	}

	var constraint TableConstraint
	if err := json.Unmarshal([]byte(`{"id":4,"tableId":1,"constraintName":"users_email_key","constraintType":"UNIQUE","columnIds":[3]}`), &constraint); err != nil {
		t.Fatalf("Unmarshal minimal TableConstraint returned error: %v", err)
	}
	if constraint.ID != 4 || constraint.TableID != 1 || constraint.ConstraintName != "users_email_key" || constraint.ConstraintType != TableConstraintTypeUnique || !reflect.DeepEqual(constraint.ColumnIDs, []int64{3}) {
		t.Fatalf("decoded minimal constraint = %#v, want stable required fields", constraint)
	}
	if !constraint.CreatedAt.IsZero() {
		t.Fatalf("missing constraint createdAt should decode to zero value, got %v", constraint.CreatedAt)
	}
}

func TestTableFieldConstraintKnownEnumsSerializeStableStrings(t *testing.T) {
	constraintTypes := map[TableConstraintType]string{
		TableConstraintTypePrimary: "PRIMARY",
		TableConstraintTypeUnique:  "UNIQUE",
	}
	for enumValue, expected := range constraintTypes {
		if got := enumValue.String(); got != expected {
			t.Fatalf("TableConstraintType.String() = %q, want %q", got, expected)
		}
		encoded, err := json.Marshal(enumValue)
		if err != nil {
			t.Fatalf("Marshal(%q) returned error: %v", enumValue, err)
		}
		if string(encoded) != `"`+expected+`"` {
			t.Fatalf("Marshal(%q) = %s, want quoted stable enum string %q", enumValue, encoded, expected)
		}
		var decoded TableConstraintType
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("Unmarshal(%s) returned error: %v", encoded, err)
		}
		if decoded != enumValue || !decoded.IsKnown() {
			t.Fatalf("decoded constraint type = %q known=%v, want %q known", decoded, decoded.IsKnown(), enumValue)
		}
	}

	logicalKinds := map[ColumnLogicalKind]string{
		ColumnLogicalKindUnknown:  "unknown",
		ColumnLogicalKindString:   "string",
		ColumnLogicalKindText:     "text",
		ColumnLogicalKindInteger:  "integer",
		ColumnLogicalKindDecimal:  "decimal",
		ColumnLogicalKindFloat:    "float",
		ColumnLogicalKindBoolean:  "boolean",
		ColumnLogicalKindDate:     "date",
		ColumnLogicalKindTime:     "time",
		ColumnLogicalKindDateTime: "datetime",
		ColumnLogicalKindBinary:   "binary",
		ColumnLogicalKindJSON:     "json",
		ColumnLogicalKindUUID:     "uuid",
		ColumnLogicalKindArray:    "array",
		ColumnLogicalKindEnum:     "enum",
	}
	for enumValue, expected := range logicalKinds {
		if got := enumValue.String(); got != expected {
			t.Fatalf("ColumnLogicalKind.String() = %q, want %q", got, expected)
		}
		encoded, err := json.Marshal(enumValue)
		if err != nil {
			t.Fatalf("Marshal(%q) returned error: %v", enumValue, err)
		}
		if string(encoded) != `"`+expected+`"` {
			t.Fatalf("Marshal(%q) = %s, want quoted stable enum string %q", enumValue, encoded, expected)
		}
		var decoded ColumnLogicalKind
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("Unmarshal(%s) returned error: %v", encoded, err)
		}
		if decoded != enumValue || !decoded.IsKnown() {
			t.Fatalf("decoded logical kind = %q known=%v, want %q known", decoded, decoded.IsKnown(), enumValue)
		}
	}
}

func TestValidationBoundaryCoversBasicMultiErrorUpstreamReferencesAndOutOfScopeLimits(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(-time.Second)
	zeroScannedAt := time.Time{}

	tableIssues := ValidateTable(DbTable{
		ID:        -1,
		SchemaID:  0,
		TableName: "bad/table",
		ScannedAt: &zeroScannedAt,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, SchemaValidationModeDraft)
	assertIssuePaths(t, tableIssues, []string{"id", "schemaId", "tableName", "scannedAt", "updatedAt"})
	assertIssueCodes(t, tableIssues, map[string]SchemaIssueCode{
		"id":        SchemaIssueCodeInvalidID,
		"schemaId":  SchemaIssueCodeInvalidID,
		"tableName": SchemaIssueCodeInvalidName,
		"scannedAt": SchemaIssueCodeInvalidTime,
		"updatedAt": SchemaIssueCodeInvalidTime,
	})
	assertAllIssuesSafeErrors(t, tableIssues)

	columnIssues := ValidateColumn(DbColumn{
		ID:              -1,
		TableID:         -10,
		OrdinalPosition: -1,
		ColumnName:      "",
		NativeType:      "\t",
		LogicalType: ColumnLogicalType{
			Kind:       ColumnLogicalKindUnknown,
			NativeType: " ",
		},
	}, SchemaValidationModeDraft)
	assertIssuePaths(t, columnIssues, []string{"id", "tableId", "ordinalPosition", "columnName", "nativeType", "logicalType.nativeType"})
	assertIssueCodes(t, columnIssues, map[string]SchemaIssueCode{
		"id":                     SchemaIssueCodeInvalidID,
		"tableId":                SchemaIssueCodeInvalidID,
		"ordinalPosition":        SchemaIssueCodeValidationFailed,
		"columnName":             SchemaIssueCodeRequired,
		"nativeType":             SchemaIssueCodeRequired,
		"logicalType.nativeType": SchemaIssueCodeRequired,
	})
	assertAllIssuesSafeErrors(t, columnIssues)

	constraintIssues := ValidateConstraint(TableConstraint{
		ID:             0,
		TableID:        10,
		ConstraintName: "users_future_key",
		ConstraintType: TableConstraintType("FOREIGN_KEY"),
		ColumnIDs:      []int64{20, -1, 20, 0},
	}, SchemaValidationModePersisted)
	assertIssuePaths(t, constraintIssues, []string{"id", "constraintType", "columnIds", "columnIds", "createdAt"})
	assertAllIssuesSafeErrors(t, constraintIssues)
	if len(constraintIssues) != 5 {
		t.Fatalf("ValidateConstraint should return all diagnosable boundary issues, got %#v", constraintIssues)
	}
	assertNoIssuePath(t, constraintIssues, "foreignKey")
	assertNoIssuePath(t, constraintIssues, "referencedTableId")
}

func TestValidationBoundaryDoesNotImplementAggregateOrDatabaseBackedReferenceChecks(t *testing.T) {
	constraint := TableConstraint{
		TableID:        10,
		ConstraintName: "users_email_key",
		ConstraintType: TableConstraintTypeUnique,
		ColumnIDs:      []int64{9999},
	}
	if issues := ValidateConstraint(constraint, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateConstraint should not verify column IDs against database or aggregate state, got %#v", issues)
	}

	column := DbColumn{
		TableID:         9999,
		OrdinalPosition: 1,
		ColumnName:      "email",
		NativeType:      "varchar(255)",
		LogicalType:     ColumnLogicalType{Kind: ColumnLogicalKindString},
	}
	if issues := ValidateColumn(column, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateColumn should only require a positive tableId and not load parent table state, got %#v", issues)
	}

	table := DbTable{SchemaID: 9999, TableName: "users"}
	if issues := ValidateTable(table, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateTable should only require a positive schemaId and not load parent schema state, got %#v", issues)
	}
}

func TestTableFieldConstraintBoundaryDoesNotImportOrExposeOutOfScopeFeatures(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeOf(DbTable{}), reflect.TypeOf(DbColumn{}), reflect.TypeOf(TableConstraint{}), reflect.TypeOf(ColumnLogicalType{})} {
		for index := range typ.NumField() {
			field := typ.Field(index)
			fieldName := strings.ToLower(field.Name)
			jsonName := strings.ToLower(strings.Split(field.Tag.Get("json"), ",")[0])

			for _, forbidden := range []string{"service", "api", "ui", "wails", "vue", "execution", "engine", "driver", "foreign", "relation", "project", "check", "index"} {
				if strings.Contains(fieldName, forbidden) || strings.Contains(jsonName, forbidden) {
					t.Fatalf("%s.%s exposes out-of-scope field matching %q with json tag %q", typ.Name(), field.Name, forbidden, field.Tag.Get("json"))
				}
			}
		}
	}

	for _, file := range schemaProductionFiles(t) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse imports for %s: %v", file, err)
		}

		for _, importSpec := range parsed.Imports {
			importPath := strings.Trim(importSpec.Path.Value, "\"")
			for _, forbidden := range []string{
				"github.com/gerdong/loomidbx/internal/dbx",
				"github.com/gerdong/loomidbx/internal/service",
				"github.com/gerdong/loomidbx/internal/repository",
				"github.com/gerdong/loomidbx/internal/storage",
				"github.com/wailsapp/wails",
				"database/sql",
			} {
				if importPath == forbidden || strings.HasPrefix(importPath, forbidden+"/") {
					t.Fatalf("schema domain file %s imports out-of-scope package %q", file, importPath)
				}
			}
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

func assertNoIssuePath(t *testing.T, issues []SchemaValidationIssue, forbidden string) {
	t.Helper()

	for _, issue := range issues {
		if issue.Path == forbidden {
			t.Fatalf("unexpected issue path %q in %#v", forbidden, issues)
		}
	}
}
