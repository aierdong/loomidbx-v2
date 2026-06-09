package schema

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
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

func TestCoreSchemaModelsExposeOnlyStableDesignFields(t *testing.T) {
	assertStructJSONFieldSet(t, reflect.TypeOf(DbCatalog{}), []string{"id", "connectionId", "catalogName", "scannedAt", "createdAt", "updatedAt"})
	assertStructJSONFieldSet(t, reflect.TypeOf(DbSchema{}), []string{"id", "catalogId", "schemaName", "scannedAt", "createdAt", "updatedAt"})
	assertStructJSONFieldSet(t, reflect.TypeOf(SchemaIdentity{}), []string{"connectionId", "catalogName", "schemaName"})
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
