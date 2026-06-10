package schema

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSchemaDomainScaffoldExportsStableShapes(t *testing.T) {
	if got, want := reflect.TypeOf(DbCatalog{}).PkgPath(), "github.com/gerdong/loomidbx/internal/domain/schema"; got != want {
		t.Fatalf("DbCatalog package path = %q, want %q", got, want)
	}

	assertJSONTags(t, reflect.TypeOf(DbCatalog{}), map[string]string{
		"ID":           "id",
		"ConnectionID": "connectionId",
		"CatalogName":  "catalogName",
		"ScannedAt":    "scannedAt",
		"CreatedAt":    "createdAt",
		"UpdatedAt":    "updatedAt",
	})
	assertJSONTags(t, reflect.TypeOf(DbSchema{}), map[string]string{
		"ID":         "id",
		"CatalogID":  "catalogId",
		"SchemaName": "schemaName",
		"ScannedAt":  "scannedAt",
		"CreatedAt":  "createdAt",
		"UpdatedAt":  "updatedAt",
	})
	assertJSONTags(t, reflect.TypeOf(SchemaIdentity{}), map[string]string{
		"ConnectionID": "connectionId",
		"CatalogName":  "catalogName",
		"SchemaName":   "schemaName",
	})
	assertJSONTags(t, reflect.TypeOf(SchemaValidationIssue{}), map[string]string{
		"Path":     "path",
		"Code":     "code",
		"Severity": "severity",
		"Message":  "message",
	})
}

func TestSchemaDomainScaffoldDeclaresStableEnums(t *testing.T) {
	tests := map[string]string{
		"SchemaIssueSeverityInfo":         string(SchemaIssueSeverityInfo),
		"SchemaIssueSeverityWarning":      string(SchemaIssueSeverityWarning),
		"SchemaIssueSeverityError":        string(SchemaIssueSeverityError),
		"SchemaIssueCodeValidationFailed": string(SchemaIssueCodeValidationFailed),
		"SchemaIssueCodeRequired":         string(SchemaIssueCodeRequired),
		"SchemaIssueCodeInvalidID":        string(SchemaIssueCodeInvalidID),
		"SchemaIssueCodeInvalidName":      string(SchemaIssueCodeInvalidName),
		"SchemaIssueCodeInvalidTime":      string(SchemaIssueCodeInvalidTime),
		"SchemaIssueCodeInvalidIdentity":  string(SchemaIssueCodeInvalidIdentity),
		"SchemaValidationModeDraft":       string(SchemaValidationModeDraft),
		"SchemaValidationModePersisted":   string(SchemaValidationModePersisted),
	}

	expected := map[string]string{
		"SchemaIssueSeverityInfo":         "info",
		"SchemaIssueSeverityWarning":      "warning",
		"SchemaIssueSeverityError":        "error",
		"SchemaIssueCodeValidationFailed": "VALIDATION_FAILED",
		"SchemaIssueCodeRequired":         "SCHEMA_REQUIRED",
		"SchemaIssueCodeInvalidID":        "SCHEMA_INVALID_ID",
		"SchemaIssueCodeInvalidName":      "SCHEMA_INVALID_NAME",
		"SchemaIssueCodeInvalidTime":      "SCHEMA_INVALID_TIME",
		"SchemaIssueCodeInvalidIdentity":  "SCHEMA_INVALID_IDENTITY",
		"SchemaValidationModeDraft":       "draft",
		"SchemaValidationModePersisted":   "persisted",
	}

	for name, got := range tests {
		if got != expected[name] {
			t.Fatalf("%s = %q, want %q", name, got, expected[name])
		}
	}
}

func TestDbSchemaImplicitSchemaNameRoundTripRequiresPresentEmptyString(t *testing.T) {
	original := DbSchema{ID: 11, CatalogID: 22, SchemaName: ""}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal(DbSchema) returned error: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("Unmarshal encoded DbSchema into field map returned error: %v", err)
	}
	assertJSONFieldsPresent(t, fields, "id", "catalogId", "schemaName", "scannedAt", "createdAt", "updatedAt")
	assertJSONStringField(t, fields, "schemaName", "")

	var decoded DbSchema
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(DbSchema) with implicit schema returned error: %v", err)
	}
	if decoded.SchemaName != "" {
		t.Fatalf("decoded SchemaName = %q, want empty implicit schema", decoded.SchemaName)
	}
	if decoded.ID != original.ID || decoded.CatalogID != original.CatalogID {
		t.Fatalf("decoded identity fields = (%d, %d), want (%d, %d)", decoded.ID, decoded.CatalogID, original.ID, original.CatalogID)
	}
}

func TestDbSchemaUnmarshalRejectsMissingAndNullSchemaName(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "missing", json: `{"id":1,"catalogId":2}`},
		{name: "null", json: `{"id":1,"catalogId":2,"schemaName":null}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var decoded DbSchema
			err := json.Unmarshal([]byte(tt.json), &decoded)
			if err == nil {
				t.Fatalf("Unmarshal(DbSchema) should reject %s schemaName", tt.name)
			}
			assertSchemaJSONError(t, err, "schemaName", SchemaIssueCodeRequired, SchemaIssueSeverityError)
		})
	}
}

func TestSchemaIdentityImplicitSchemaNameRoundTripRequiresPresentEmptyString(t *testing.T) {
	original := SchemaIdentity{ConnectionID: 44, CatalogName: "analytics", SchemaName: ""}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal(SchemaIdentity) returned error: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("Unmarshal encoded SchemaIdentity into field map returned error: %v", err)
	}
	assertJSONFieldsPresent(t, fields, "connectionId", "catalogName", "schemaName")
	assertJSONStringField(t, fields, "schemaName", "")

	var decoded SchemaIdentity
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(SchemaIdentity) with implicit schema returned error: %v", err)
	}
	if decoded.SchemaName != "" {
		t.Fatalf("decoded SchemaName = %q, want empty implicit schema", decoded.SchemaName)
	}
	if decoded.ConnectionID != original.ConnectionID || decoded.CatalogName != original.CatalogName {
		t.Fatalf("decoded identity = (%d, %q), want (%d, %q)", decoded.ConnectionID, decoded.CatalogName, original.ConnectionID, original.CatalogName)
	}
}

func TestSchemaIdentityUnmarshalRejectsMissingAndNullSchemaName(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "missing", json: `{"connectionId":1,"catalogName":"analytics"}`},
		{name: "null", json: `{"connectionId":1,"catalogName":"analytics","schemaName":null}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var decoded SchemaIdentity
			err := json.Unmarshal([]byte(tt.json), &decoded)
			if err == nil {
				t.Fatalf("Unmarshal(SchemaIdentity) should reject %s schemaName", tt.name)
			}
			assertSchemaJSONError(t, err, "schemaName", SchemaIssueCodeRequired, SchemaIssueSeverityError)
		})
	}
}

func TestCoreTableColumnModelsExposeStableDesignFields(t *testing.T) {
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

	assertStructJSONFieldSet(t, reflect.TypeOf(DbTable{}), []string{"id", "schemaId", "tableName", "comment", "ddlSnapshot", "scannedAt", "createdAt", "updatedAt"})
	assertStructJSONFieldSet(t, reflect.TypeOf(DbColumn{}), []string{"id", "tableId", "ordinalPosition", "columnName", "nativeType", "logicalType", "nullable", "defaultValue", "isPrimaryKey", "comment", "createdAt", "updatedAt"})
}

func TestTableColumnSerializationPreservesStableIdentityParentsAndNullableFields(t *testing.T) {
	scannedAt := time.Date(2026, 6, 10, 9, 30, 0, 0, time.UTC)
	createdAt := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 10, 11, 0, 0, 0, time.UTC)
	defaultValue := "CURRENT_TIMESTAMP"

	table := DbTable{
		ID:          401,
		SchemaID:    303,
		TableName:   "orders",
		Comment:     "business order table",
		DDLSnapshot: "CREATE TABLE orders (...) ",
		ScannedAt:   &scannedAt,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
	assertJSONRoundTrip(t, "DbTable", table)

	column := DbColumn{
		ID:              501,
		TableID:         table.ID,
		OrdinalPosition: 2,
		ColumnName:      "created_at",
		NativeType:      "timestamp with time zone",
		LogicalType:     ColumnLogicalType{},
		Nullable:        false,
		DefaultValue:    &defaultValue,
		IsPrimaryKey:    false,
		Comment:         "creation timestamp",
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
	assertJSONRoundTrip(t, "DbColumn", column)

	encoded, err := json.Marshal(DbColumn{ID: 502, TableID: table.ID, OrdinalPosition: 3, ColumnName: "description", NativeType: "text", LogicalType: ColumnLogicalType{}, Nullable: true})
	if err != nil {
		t.Fatalf("Marshal(DbColumn) with nil defaultValue returned error: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("Unmarshal encoded DbColumn into field map returned error: %v", err)
	}
	assertJSONFieldsPresent(t, fields, "id", "tableId", "ordinalPosition", "columnName", "nativeType", "logicalType", "nullable", "defaultValue", "isPrimaryKey", "comment", "createdAt", "updatedAt")
	if string(fields["defaultValue"]) != "null" {
		t.Fatalf("defaultValue JSON = %s, want null for nil default value", fields["defaultValue"])
	}
	if string(fields["nullable"]) != "true" {
		t.Fatalf("nullable JSON = %s, want true", fields["nullable"])
	}
	if string(fields["isPrimaryKey"]) != "false" {
		t.Fatalf("isPrimaryKey JSON = %s, want false", fields["isPrimaryKey"])
	}
}

func TestTableColumnModelsUseScalarParentReferencesAndNoOutOfScopeFields(t *testing.T) {
	tableType := reflect.TypeOf(DbTable{})
	assertFieldType(t, tableType, "ID", reflect.TypeOf(int64(0)))
	assertFieldType(t, tableType, "SchemaID", reflect.TypeOf(int64(0)))

	columnType := reflect.TypeOf(DbColumn{})
	assertFieldType(t, columnType, "ID", reflect.TypeOf(int64(0)))
	assertFieldType(t, columnType, "TableID", reflect.TypeOf(int64(0)))
	assertFieldType(t, columnType, "OrdinalPosition", reflect.TypeOf(int(0)))
	assertFieldType(t, columnType, "LogicalType", reflect.TypeOf(ColumnLogicalType{}))
	assertFieldType(t, columnType, "DefaultValue", reflect.TypeOf((*string)(nil)))

	for _, typ := range []reflect.Type{tableType, columnType} {
		for index := range typ.NumField() {
			field := typ.Field(index)
			fieldName := strings.ToLower(field.Name)
			jsonName := strings.ToLower(strings.Split(field.Tag.Get("json"), ",")[0])

			for _, forbidden := range []string{"service", "api", "ui", "wails", "vue", "execution", "engine", "driver", "sql", "foreign", "relation", "project"} {
				if strings.Contains(fieldName, forbidden) || strings.Contains(jsonName, forbidden) {
					t.Fatalf("%s.%s exposes out-of-scope field matching %q with json tag %q", typ.Name(), field.Name, forbidden, field.Tag.Get("json"))
				}
			}
		}
	}
}

func TestCoreSchemaModelsExposeOnlyStableDesignFields(t *testing.T) {
	assertStructJSONFieldSet(t, reflect.TypeOf(DbCatalog{}), []string{"id", "connectionId", "catalogName", "scannedAt", "createdAt", "updatedAt"})
	assertStructJSONFieldSet(t, reflect.TypeOf(DbSchema{}), []string{"id", "catalogId", "schemaName", "scannedAt", "createdAt", "updatedAt"})
	assertStructJSONFieldSet(t, reflect.TypeOf(SchemaIdentity{}), []string{"connectionId", "catalogName", "schemaName"})
}

func TestSchemaSerializationRoundTripPreservesCatalogSchemaIdentityAndIssueContracts(t *testing.T) {
	scannedAt := time.Date(2026, 6, 10, 9, 30, 0, 0, time.UTC)
	createdAt := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 10, 11, 0, 0, 0, time.UTC)

	catalog := DbCatalog{
		ID:           101,
		ConnectionID: 202,
		CatalogName:  "analytics",
		ScannedAt:    &scannedAt,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
	assertJSONRoundTrip(t, "DbCatalog", catalog)

	schema := DbSchema{
		ID:         303,
		CatalogID:  101,
		SchemaName: "public",
		ScannedAt:  &scannedAt,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
	assertJSONRoundTrip(t, "DbSchema", schema)

	identity := SchemaIdentity{ConnectionID: 202, CatalogName: "analytics", SchemaName: "public"}
	assertJSONRoundTrip(t, "SchemaIdentity", identity)

	issue := SchemaValidationIssue{
		Path:     "identity.schemaName",
		Code:     SchemaIssueCodeRequired,
		Severity: SchemaIssueSeverityError,
		Message:  "schemaName is required",
	}
	assertJSONRoundTrip(t, "SchemaValidationIssue", issue)
}

func TestSchemaJSONMissingOptionalFieldsDecodeToCompatibleZeroValues(t *testing.T) {
	var catalog DbCatalog
	if err := json.Unmarshal([]byte(`{"id":1,"connectionId":2,"catalogName":"analytics"}`), &catalog); err != nil {
		t.Fatalf("Unmarshal minimal DbCatalog returned error: %v", err)
	}
	if catalog.ID != 1 || catalog.ConnectionID != 2 || catalog.CatalogName != "analytics" {
		t.Fatalf("decoded minimal catalog = %#v, want stable required fields", catalog)
	}
	if catalog.ScannedAt != nil {
		t.Fatalf("missing scannedAt should decode to nil, got %#v", catalog.ScannedAt)
	}
	if !catalog.CreatedAt.IsZero() || !catalog.UpdatedAt.IsZero() {
		t.Fatalf("missing audit times should decode to zero values, got createdAt=%v updatedAt=%v", catalog.CreatedAt, catalog.UpdatedAt)
	}

	var schema DbSchema
	if err := json.Unmarshal([]byte(`{"id":3,"catalogId":1,"schemaName":""}`), &schema); err != nil {
		t.Fatalf("Unmarshal minimal implicit DbSchema returned error: %v", err)
	}
	if schema.ID != 3 || schema.CatalogID != 1 || schema.SchemaName != "" {
		t.Fatalf("decoded minimal schema = %#v, want stable identity fields with implicit schema", schema)
	}
	if schema.ScannedAt != nil {
		t.Fatalf("missing schema scannedAt should decode to nil, got %#v", schema.ScannedAt)
	}
	if !schema.CreatedAt.IsZero() || !schema.UpdatedAt.IsZero() {
		t.Fatalf("missing schema audit times should decode to zero values, got createdAt=%v updatedAt=%v", schema.CreatedAt, schema.UpdatedAt)
	}
}

func TestSchemaUnmarshalSchemaNameErrorsUseFieldLevelIssueShape(t *testing.T) {
	tests := []struct {
		name    string
		decode  func() error
		path    string
		message string
	}{
		{
			name: "DbSchema missing schemaName",
			decode: func() error {
				var decoded DbSchema
				return json.Unmarshal([]byte(`{"id":1,"catalogId":2}`), &decoded)
			},
			path:    "schemaName",
			message: "schemaName is required and must be a string; use an empty string for an implicit schema",
		},
		{
			name: "DbSchema null schemaName",
			decode: func() error {
				var decoded DbSchema
				return json.Unmarshal([]byte(`{"id":1,"catalogId":2,"schemaName":null}`), &decoded)
			},
			path:    "schemaName",
			message: "schemaName is required and must be a string; use an empty string for an implicit schema",
		},
		{
			name: "SchemaIdentity missing schemaName",
			decode: func() error {
				var decoded SchemaIdentity
				return json.Unmarshal([]byte(`{"connectionId":1,"catalogName":"analytics"}`), &decoded)
			},
			path:    "schemaName",
			message: "schemaName is required and must be a string; use an empty string for an implicit schema",
		},
		{
			name: "SchemaIdentity null schemaName",
			decode: func() error {
				var decoded SchemaIdentity
				return json.Unmarshal([]byte(`{"connectionId":1,"catalogName":"analytics","schemaName":null}`), &decoded)
			},
			path:    "schemaName",
			message: "schemaName is required and must be a string; use an empty string for an implicit schema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.decode()
			if err == nil {
				t.Fatalf("Unmarshal should reject %s", tt.name)
			}

			var schemaErr SchemaJSONError
			if !errors.As(err, &schemaErr) {
				t.Fatalf("error type = %T, want SchemaJSONError", err)
			}
			issues := schemaErr.Issues()
			if got, want := len(issues), 1; got != want {
				t.Fatalf("SchemaJSONError issues = %#v, want exactly %d", issues, want)
			}
			expected := SchemaValidationIssue{Path: tt.path, Code: SchemaIssueCodeRequired, Severity: SchemaIssueSeverityError, Message: tt.message}
			if !reflect.DeepEqual(issues[0], expected) {
				t.Fatalf("SchemaJSONError issue = %#v, want %#v", issues[0], expected)
			}

			encoded, err := json.Marshal(issues[0])
			if err != nil {
				t.Fatalf("Marshal SchemaJSONError issue returned error: %v", err)
			}
			const expectedFields = 4
			var fieldShape map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fieldShape); err != nil {
				t.Fatalf("Unmarshal encoded issue shape returned error: %v", err)
			}
			if got := len(fieldShape); got != expectedFields {
				t.Fatalf("issue JSON field count = %d, want %d: %s", got, expectedFields, encoded)
			}
			assertJSONFieldsPresent(t, fieldShape, "path", "code", "severity", "message")
		})
	}
}

func TestSchemaEnumSerializationRejectsNonStringJSONValues(t *testing.T) {
	tests := []struct {
		name   string
		decode func() error
	}{
		{
			name: "issue code object",
			decode: func() error {
				var decoded SchemaIssueCode
				return json.Unmarshal([]byte(`{"code":"SCHEMA_REQUIRED"}`), &decoded)
			},
		},
		{
			name: "severity number",
			decode: func() error {
				var decoded SchemaIssueSeverity
				return json.Unmarshal([]byte(`1`), &decoded)
			},
		},
		{
			name: "validation mode boolean",
			decode: func() error {
				var decoded SchemaValidationMode
				return json.Unmarshal([]byte(`true`), &decoded)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.decode(); err == nil {
				t.Fatalf("enum unmarshal should reject non-string JSON for %s", tt.name)
			}
		})
	}
}

func TestSchemaNamePresentAsNullNeverDecodesToImplicitSchema(t *testing.T) {
	var decodedSchema DbSchema
	if err := json.Unmarshal([]byte(`{"id":1,"catalogId":2,"schemaName":null}`), &decodedSchema); err == nil {
		t.Fatalf("DbSchema null schemaName should return a field-level error, not decode to implicit schema: %#v", decodedSchema)
	}
	if decodedSchema.SchemaName == "" && decodedSchema.CatalogID == 2 {
		t.Fatalf("DbSchema null schemaName partially decoded as implicit schema: %#v", decodedSchema)
	}

	var decodedIdentity SchemaIdentity
	if err := json.Unmarshal([]byte(`{"connectionId":1,"catalogName":"analytics","schemaName":null}`), &decodedIdentity); err == nil {
		t.Fatalf("SchemaIdentity null schemaName should return a field-level error, not decode to implicit schema: %#v", decodedIdentity)
	}
	if decodedIdentity.SchemaName == "" && decodedIdentity.ConnectionID == 1 {
		t.Fatalf("SchemaIdentity null schemaName partially decoded as implicit schema: %#v", decodedIdentity)
	}
}

func assertJSONRoundTrip[T any](t *testing.T, name string, original T) {
	t.Helper()

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal(%s) returned error: %v", name, err)
	}

	var decoded T
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(%s) returned error: %v", name, err)
	}
	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("%s round trip = %#v, want %#v", name, decoded, original)
	}
}

func assertJSONTags(t *testing.T, typ reflect.Type, expected map[string]string) {
	t.Helper()

	for fieldName, jsonName := range expected {
		field, ok := typ.FieldByName(fieldName)
		if !ok {
			t.Fatalf("%s missing field %s", typ.Name(), fieldName)
		}

		tag := field.Tag.Get("json")
		if tag == "" {
			t.Fatalf("%s.%s missing json tag", typ.Name(), fieldName)
		}
		actualName := strings.Split(tag, ",")[0]
		if actualName != jsonName {
			t.Fatalf("%s.%s json tag = %q, want %q", typ.Name(), fieldName, actualName, jsonName)
		}
		if jsonName == "schemaName" && strings.Contains(tag, "omitempty") {
			t.Fatalf("%s.%s must not use omitempty so implicit schema serializes as empty string", typ.Name(), fieldName)
		}
	}
}

func assertStructJSONFieldSet(t *testing.T, typ reflect.Type, expected []string) {
	t.Helper()

	if typ.NumField() != len(expected) {
		t.Fatalf("%s field count = %d, want %d", typ.Name(), typ.NumField(), len(expected))
	}

	actual := make([]string, 0, typ.NumField())
	for i := range typ.NumField() {
		field := typ.Field(i)
		actual = append(actual, strings.Split(field.Tag.Get("json"), ",")[0])
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%s json fields = %#v, want %#v", typ.Name(), actual, expected)
	}
}

func assertJSONFieldsPresent(t *testing.T, fields map[string]json.RawMessage, expected ...string) {
	t.Helper()

	for _, field := range expected {
		if _, ok := fields[field]; !ok {
			t.Fatalf("encoded JSON missing stable field %q in %#v", field, fields)
		}
	}
}

func assertJSONStringField(t *testing.T, fields map[string]json.RawMessage, field string, expected string) {
	t.Helper()

	raw, ok := fields[field]
	if !ok {
		t.Fatalf("encoded JSON missing field %q", field)
	}
	var actual string
	if err := json.Unmarshal(raw, &actual); err != nil {
		t.Fatalf("field %q should be a JSON string: %v", field, err)
	}
	if actual != expected {
		t.Fatalf("field %q = %q, want %q", field, actual, expected)
	}
}

func assertSchemaJSONError(t *testing.T, err error, path string, code SchemaIssueCode, severity SchemaIssueSeverity) {
	t.Helper()

	var schemaErr SchemaJSONError
	if !errors.As(err, &schemaErr) {
		t.Fatalf("error type = %T, want SchemaJSONError: %v", err, err)
	}

	issues := schemaErr.Issues()
	if len(issues) != 1 {
		t.Fatalf("SchemaJSONError issues = %#v, want exactly one", issues)
	}
	issue := issues[0]
	if issue.Path != path || issue.Code != code || issue.Severity != severity {
		t.Fatalf("SchemaJSONError issue = %#v, want path=%q code=%q severity=%q", issue, path, code, severity)
	}
	if strings.TrimSpace(issue.Message) == "" {
		t.Fatalf("SchemaJSONError issue should include a safe message: %#v", issue)
	}
}
