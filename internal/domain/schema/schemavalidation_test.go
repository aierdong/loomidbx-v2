package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSchemaIssueCodeStableStringValuesAndRecognition(t *testing.T) {
	tests := []struct {
		name     string
		code     SchemaIssueCode
		expected string
	}{
		{name: "validation failed", code: SchemaIssueCodeValidationFailed, expected: "VALIDATION_FAILED"},
		{name: "required", code: SchemaIssueCodeRequired, expected: "SCHEMA_REQUIRED"},
		{name: "invalid id", code: SchemaIssueCodeInvalidID, expected: "SCHEMA_INVALID_ID"},
		{name: "invalid name", code: SchemaIssueCodeInvalidName, expected: "SCHEMA_INVALID_NAME"},
		{name: "invalid time", code: SchemaIssueCodeInvalidTime, expected: "SCHEMA_INVALID_TIME"},
		{name: "invalid identity", code: SchemaIssueCodeInvalidIdentity, expected: "SCHEMA_INVALID_IDENTITY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.code.IsKnown() {
				t.Fatalf("%s should be recognized as a known schema issue code", tt.code)
			}
			if tt.code.IsUnknown() {
				t.Fatalf("%s should not be recognized as unknown", tt.code)
			}
			if got := tt.code.String(); got != tt.expected {
				t.Fatalf("String() = %q, want %q", got, tt.expected)
			}
		})
	}

	unknown := SchemaIssueCode("SCHEMA_FUTURE_CODE")
	if unknown.IsKnown() {
		t.Fatalf("unknown schema issue code %q should not be known", unknown)
	}
	if !unknown.IsUnknown() {
		t.Fatalf("unknown schema issue code %q should be explicitly unknown", unknown)
	}
}

func TestSchemaIssueSeverityStableStringValuesAndRecognition(t *testing.T) {
	tests := []struct {
		name     string
		severity SchemaIssueSeverity
		expected string
	}{
		{name: "info", severity: SchemaIssueSeverityInfo, expected: "info"},
		{name: "warning", severity: SchemaIssueSeverityWarning, expected: "warning"},
		{name: "error", severity: SchemaIssueSeverityError, expected: "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.severity.IsKnown() {
				t.Fatalf("%s should be recognized as a known schema issue severity", tt.severity)
			}
			if tt.severity.IsUnknown() {
				t.Fatalf("%s should not be recognized as unknown", tt.severity)
			}
			if got := tt.severity.String(); got != tt.expected {
				t.Fatalf("String() = %q, want %q", got, tt.expected)
			}
		})
	}

	unknown := SchemaIssueSeverity("fatal")
	if unknown.IsKnown() {
		t.Fatalf("unknown schema issue severity %q should not be known", unknown)
	}
	if !unknown.IsUnknown() {
		t.Fatalf("unknown schema issue severity %q should be explicitly unknown", unknown)
	}
}

func TestSchemaValidationModeStableStringValuesAndRecognition(t *testing.T) {
	tests := []struct {
		name     string
		mode     SchemaValidationMode
		expected string
	}{
		{name: "draft", mode: SchemaValidationModeDraft, expected: "draft"},
		{name: "persisted", mode: SchemaValidationModePersisted, expected: "persisted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.mode.IsKnown() {
				t.Fatalf("%s should be recognized as a known schema validation mode", tt.mode)
			}
			if tt.mode.IsUnknown() {
				t.Fatalf("%s should not be recognized as unknown", tt.mode)
			}
			if got := tt.mode.String(); got != tt.expected {
				t.Fatalf("String() = %q, want %q", got, tt.expected)
			}
		})
	}

	unknown := SchemaValidationMode("runtime")
	if unknown.IsKnown() {
		t.Fatalf("unknown schema validation mode %q should not be known", unknown)
	}
	if !unknown.IsUnknown() {
		t.Fatalf("unknown schema validation mode %q should be explicitly unknown", unknown)
	}
}

func TestSchemaEnumsJSONRoundTripPreservesKnownAndUnknownValues(t *testing.T) {
	tests := []struct {
		name          string
		value         any
		jsonValue     string
		expectedKnown bool
		decode        func([]byte) (string, bool)
	}{
		{
			name:          "known issue code",
			value:         SchemaIssueCodeRequired,
			jsonValue:     `"SCHEMA_REQUIRED"`,
			expectedKnown: true,
			decode: func(data []byte) (string, bool) {
				var decoded SchemaIssueCode
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("Unmarshal SchemaIssueCode returned error: %v", err)
				}
				return decoded.String(), decoded.IsKnown()
			},
		},
		{
			name:          "unknown issue code",
			value:         SchemaIssueCode("SCHEMA_FUTURE_CODE"),
			jsonValue:     `"SCHEMA_FUTURE_CODE"`,
			expectedKnown: false,
			decode: func(data []byte) (string, bool) {
				var decoded SchemaIssueCode
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("Unmarshal SchemaIssueCode returned error: %v", err)
				}
				return decoded.String(), decoded.IsKnown()
			},
		},
		{
			name:          "known severity",
			value:         SchemaIssueSeverityError,
			jsonValue:     `"error"`,
			expectedKnown: true,
			decode: func(data []byte) (string, bool) {
				var decoded SchemaIssueSeverity
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("Unmarshal SchemaIssueSeverity returned error: %v", err)
				}
				return decoded.String(), decoded.IsKnown()
			},
		},
		{
			name:          "unknown severity",
			value:         SchemaIssueSeverity("fatal"),
			jsonValue:     `"fatal"`,
			expectedKnown: false,
			decode: func(data []byte) (string, bool) {
				var decoded SchemaIssueSeverity
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("Unmarshal SchemaIssueSeverity returned error: %v", err)
				}
				return decoded.String(), decoded.IsKnown()
			},
		},
		{
			name:          "known validation mode",
			value:         SchemaValidationModeDraft,
			jsonValue:     `"draft"`,
			expectedKnown: true,
			decode: func(data []byte) (string, bool) {
				var decoded SchemaValidationMode
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("Unmarshal SchemaValidationMode returned error: %v", err)
				}
				return decoded.String(), decoded.IsKnown()
			},
		},
		{
			name:          "unknown validation mode",
			value:         SchemaValidationMode("runtime"),
			jsonValue:     `"runtime"`,
			expectedKnown: false,
			decode: func(data []byte) (string, bool) {
				var decoded SchemaValidationMode
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("Unmarshal SchemaValidationMode returned error: %v", err)
				}
				return decoded.String(), decoded.IsKnown()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal(%v) returned error: %v", tt.value, err)
			}
			if string(encoded) != tt.jsonValue {
				t.Fatalf("Marshal(%v) = %s, want %s", tt.value, encoded, tt.jsonValue)
			}

			decoded, known := tt.decode(encoded)
			if decoded != strings.Trim(tt.jsonValue, `\"`) {
				t.Fatalf("decoded value = %q, want %s", decoded, tt.jsonValue)
			}
			if known != tt.expectedKnown {
				t.Fatalf("decoded known = %v, want %v", known, tt.expectedKnown)
			}
		})
	}
}

func TestSchemaValidationIssueJSONShapeAndRoundTrip(t *testing.T) {
	issue := SchemaValidationIssue{
		Path:     "identity.schemaName",
		Code:     SchemaIssueCodeRequired,
		Severity: SchemaIssueSeverityError,
		Message:  "schemaName is required",
	}

	encoded, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("Marshal(SchemaValidationIssue) returned error: %v", err)
	}

	const expected = `{"path":"identity.schemaName","code":"SCHEMA_REQUIRED","severity":"error","message":"schemaName is required"}`
	if string(encoded) != expected {
		t.Fatalf("SchemaValidationIssue JSON = %s, want exact ConfigIssue/ApiIssue-compatible shape %s", encoded, expected)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("Unmarshal encoded issue into map returned error: %v", err)
	}
	if got, want := len(fields), 4; got != want {
		t.Fatalf("encoded issue has %d fields, want exactly %d: %v", got, want, fields)
	}
	for _, field := range []string{"path", "code", "severity", "message"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("encoded issue missing required field %q in %s", field, encoded)
		}
	}

	var decoded SchemaValidationIssue
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(SchemaValidationIssue) returned error: %v", err)
	}
	if !reflect.DeepEqual(decoded, issue) {
		t.Fatalf("decoded issue = %#v, want %#v", decoded, issue)
	}
	if !decoded.Code.IsKnown() || !decoded.Severity.IsKnown() {
		t.Fatalf("decoded issue should preserve recognized code and severity: %#v", decoded)
	}
}

func TestSchemaValidationIssueCanCarryUnknownEnumValuesForExplicitValidation(t *testing.T) {
	const raw = `{"path":"catalogName","code":"SCHEMA_UNKNOWN","severity":"fatal","message":"future issue"}`

	var issue SchemaValidationIssue
	if err := json.Unmarshal([]byte(raw), &issue); err != nil {
		t.Fatalf("Unmarshal issue with unknown enum values returned error: %v", err)
	}
	if issue.Code.IsKnown() {
		t.Fatalf("unknown issue code %q should not be silently recognized", issue.Code)
	}
	if issue.Severity.IsKnown() {
		t.Fatalf("unknown issue severity %q should not be silently recognized", issue.Severity)
	}

	encoded, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("Marshal issue with unknown enum values returned error: %v", err)
	}
	if string(encoded) != raw {
		t.Fatalf("unknown enum values should serialize safely and unchanged: got %s, want %s", encoded, raw)
	}
}

func TestValidateIssueAcceptsReusableFieldLevelIssueContract(t *testing.T) {
	issue := SchemaValidationIssue{
		Path:     "identity.schemaName",
		Code:     SchemaIssueCodeRequired,
		Severity: SchemaIssueSeverityError,
		Message:  "schemaName is required",
	}

	issues := ValidateIssue(issue)
	if len(issues) != 0 {
		t.Fatalf("ValidateIssue(valid issue) = %#v, want no issues", issues)
	}
}

func TestValidateIssueReturnsMultipleStructuredProblems(t *testing.T) {
	issue := SchemaValidationIssue{
		Path:     "CatalogName",
		Code:     SchemaIssueCode("SCHEMA_UNKNOWN"),
		Severity: SchemaIssueSeverity("fatal"),
		Message:  "   ",
	}

	issues := ValidateIssue(issue)
	if got, want := len(issues), 4; got != want {
		t.Fatalf("ValidateIssue(invalid issue) returned %d issues, want %d: %#v", got, want, issues)
	}

	assertIssuePaths(t, issues, []string{"path", "code", "severity", "message"})
	for _, validationIssue := range issues {
		if validationIssue.Code != SchemaIssueCodeValidationFailed {
			t.Fatalf("validation issue code = %q, want %q", validationIssue.Code, SchemaIssueCodeValidationFailed)
		}
		if validationIssue.Severity != SchemaIssueSeverityError {
			t.Fatalf("validation issue severity = %q, want %q", validationIssue.Severity, SchemaIssueSeverityError)
		}
		if strings.TrimSpace(validationIssue.Message) == "" {
			t.Fatalf("validation issue should include safe message: %#v", validationIssue)
		}
	}
}

func TestValidateIssueRequiresJSONFieldPathSemantics(t *testing.T) {
	invalidPaths := []string{"catalog_name", "catalogName.", ".catalogName", "catalogName[0]", "catalogName.SchemaName"}

	for _, path := range invalidPaths {
		t.Run(path, func(t *testing.T) {
			issues := ValidateIssue(SchemaValidationIssue{
				Path:     path,
				Code:     SchemaIssueCodeRequired,
				Severity: SchemaIssueSeverityError,
				Message:  "schema field is required",
			})
			assertIssuePaths(t, issues, []string{"path"})
		})
	}
}

func TestValidateCatalogDraftAndPersistedUseDifferentIDAndAuditTimeRules(t *testing.T) {
	catalog := DbCatalog{ConnectionID: 42, CatalogName: "analytics"}
	if issues := ValidateCatalog(catalog, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateCatalog(draft) = %#v, want no issues for zero primary key and audit times", issues)
	}

	persistedIssues := ValidateCatalog(catalog, SchemaValidationModePersisted)
	assertIssuePaths(t, persistedIssues, []string{"id", "createdAt", "updatedAt"})

	negativeID := catalog
	negativeID.ID = -1
	assertIssuePaths(t, ValidateCatalog(negativeID, SchemaValidationModeDraft), []string{"id"})
}

func TestValidateCatalogReturnsMultipleFieldLevelIssues(t *testing.T) {
	zeroScanTime := time.Time{}
	createdAt := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(-time.Minute)

	issues := ValidateCatalog(DbCatalog{
		ID:           -7,
		ConnectionID: 0,
		CatalogName:  "bad/catalog",
		ScannedAt:    &zeroScanTime,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, SchemaValidationModeDraft)

	assertIssuePaths(t, issues, []string{"id", "connectionId", "catalogName", "scannedAt", "updatedAt"})
	assertAllIssuesSafeErrors(t, issues)
}

func TestValidateSchemaAllowsImplicitSchemaNameInBothModes(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	implicitSchema := DbSchema{ID: 9, CatalogID: 22, SchemaName: "", CreatedAt: createdAt, UpdatedAt: createdAt}

	if issues := ValidateSchema(implicitSchema, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateSchema(draft implicit schema) = %#v, want no issues", issues)
	}
	if issues := ValidateSchema(implicitSchema, SchemaValidationModePersisted); len(issues) != 0 {
		t.Fatalf("ValidateSchema(persisted implicit schema) = %#v, want no issues", issues)
	}
}

func TestValidateSchemaRejectsInvalidParentNameAndAuditTimes(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(-time.Minute)

	issues := ValidateSchema(DbSchema{
		CatalogID:  0,
		SchemaName: "  ",
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, SchemaValidationModeDraft)

	assertIssuePaths(t, issues, []string{"catalogId", "schemaName", "updatedAt"})
	assertAllIssuesSafeErrors(t, issues)
}

func TestValidateIdentityAllowsImplicitSchemaAndRejectsInvalidParts(t *testing.T) {
	implicitIdentity := SchemaIdentity{ConnectionID: 42, CatalogName: "analytics", SchemaName: ""}
	if issues := ValidateIdentity(implicitIdentity); len(issues) != 0 {
		t.Fatalf("ValidateIdentity(implicit schema) = %#v, want no issues", issues)
	}

	issues := ValidateIdentity(SchemaIdentity{ConnectionID: 0, CatalogName: " ", SchemaName: "bad/schema"})
	assertIssuePaths(t, issues, []string{"connectionId", "catalogName", "schemaName"})
	assertAllIssuesSafeErrors(t, issues)
}

func TestValidationEntryPointsRejectUnknownMode(t *testing.T) {
	catalogIssues := ValidateCatalog(DbCatalog{ConnectionID: 1, CatalogName: "analytics"}, SchemaValidationMode("runtime"))
	assertIssuePaths(t, catalogIssues, []string{"mode"})

	schemaIssues := ValidateSchema(DbSchema{CatalogID: 1, SchemaName: "public"}, SchemaValidationMode("runtime"))
	assertIssuePaths(t, schemaIssues, []string{"mode"})
}

func TestInMemoryUniquenessSemanticsDoNotRequireDatabaseAccess(t *testing.T) {
	catalogIssues := ValidateCatalogUniqueness([]DbCatalog{
		{ConnectionID: 1, CatalogName: "analytics"},
		{ConnectionID: 2, CatalogName: "analytics"},
		{ConnectionID: 1, CatalogName: "analytics"},
	})
	assertIssuePaths(t, catalogIssues, []string{"catalogName"})

	schemaIssues := ValidateSchemaUniqueness([]DbSchema{
		{CatalogID: 10, SchemaName: ""},
		{CatalogID: 11, SchemaName: ""},
		{CatalogID: 10, SchemaName: ""},
	})
	assertIssuePaths(t, schemaIssues, []string{"schemaName"})
	assertAllIssuesSafeErrors(t, append(catalogIssues, schemaIssues...))
}

func assertAllIssuesSafeErrors(t *testing.T, issues []SchemaValidationIssue) {
	t.Helper()

	if len(issues) == 0 {
		t.Fatalf("expected at least one issue")
	}
	for _, issue := range issues {
		if issue.Severity != SchemaIssueSeverityError {
			t.Fatalf("issue severity = %q, want error: %#v", issue.Severity, issue)
		}
		if strings.TrimSpace(issue.Message) == "" {
			t.Fatalf("issue message should be non-empty and safe: %#v", issue)
		}
	}
}

func assertIssuePaths(t *testing.T, issues []SchemaValidationIssue, expected []string) {
	t.Helper()

	actual := make([]string, 0, len(issues))
	for _, issue := range issues {
		actual = append(actual, issue.Path)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("issue paths = %#v, want %#v in %#v", actual, expected, issues)
	}
}

func TestSchemaDomainDoesNotImportInternalConfig(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(".", "*.go"))
	if err != nil {
		t.Fatalf("glob schema package files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read schema package file %s: %v", file, err)
		}
		if strings.Contains(string(content), "internal/config") {
			t.Fatalf("schema domain file %s must not directly import internal/config", file)
		}
	}
}
