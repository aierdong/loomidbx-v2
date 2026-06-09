package schema

import (
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
