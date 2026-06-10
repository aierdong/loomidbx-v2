package rule

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gerdong/loomidbx/internal/domain/schema"
)

func TestRuleValidationContractReusesSchemaIssueTypes(t *testing.T) {
	fnType := reflect.TypeOf(ruleValidationIssue)
	if got, want := fnType.NumIn(), 3; got != want {
		t.Fatalf("ruleValidationIssue input count = %d, want %d", got, want)
	}
	if got, want := fnType.In(0), reflect.TypeOf(""); got != want {
		t.Fatalf("ruleValidationIssue path type = %s, want %s", got, want)
	}
	if got, want := fnType.In(1), reflect.TypeOf(schema.SchemaIssueCodeRequired); got != want {
		t.Fatalf("ruleValidationIssue code type = %s, want %s", got, want)
	}
	if got, want := fnType.In(2), reflect.TypeOf(""); got != want {
		t.Fatalf("ruleValidationIssue message type = %s, want %s", got, want)
	}
	if got, want := fnType.NumOut(), 1; got != want {
		t.Fatalf("ruleValidationIssue output count = %d, want %d", got, want)
	}
	if got, want := fnType.Out(0), reflect.TypeOf(schema.SchemaValidationIssue{}); got != want {
		t.Fatalf("ruleValidationIssue output type = %s, want %s", got, want)
	}
}

func TestRuleValidationContractReturnsMultipleSchemaIssues(t *testing.T) {
	issues := []schema.SchemaValidationIssue{
		ruleValidationIssue("generatorName", schema.SchemaIssueCodeRuleInvalidText, "generatorName must be a safe generator identifier"),
		ruleValidationIssue("dataMappingType", schema.SchemaIssueCodeRuleInvalidEnum, "dataMappingType must be one of the stable rule mapping values"),
		ruleValidationIssue("params", schema.SchemaIssueCodeRuleSensitiveValueNotAllowed, "params must not contain sensitive credential fields"),
	}

	if got, want := len(issues), 3; got != want {
		t.Fatalf("rule issues count = %d, want %d: %#v", got, want, issues)
	}

	assertRuleValidationIssue(t, issues[0], schema.SchemaValidationIssue{
		Path:     "generatorName",
		Code:     schema.SchemaIssueCodeRuleInvalidText,
		Severity: schema.SchemaIssueSeverityError,
		Message:  "generatorName must be a safe generator identifier",
	})
	assertRuleValidationIssue(t, issues[1], schema.SchemaValidationIssue{
		Path:     "dataMappingType",
		Code:     schema.SchemaIssueCodeRuleInvalidEnum,
		Severity: schema.SchemaIssueSeverityError,
		Message:  "dataMappingType must be one of the stable rule mapping values",
	})
	assertRuleValidationIssue(t, issues[2], schema.SchemaValidationIssue{
		Path:     "params",
		Code:     schema.SchemaIssueCodeRuleSensitiveValueNotAllowed,
		Severity: schema.SchemaIssueSeverityError,
		Message:  "params must not contain sensitive credential fields",
	})
}

func TestFieldRuleIssueCodesExtendSchemaIssueCodeContract(t *testing.T) {
	tests := []struct {
		name     string
		code     schema.SchemaIssueCode
		expected string
	}{
		{name: "invalid enum", code: schema.SchemaIssueCodeRuleInvalidEnum, expected: "RULE_INVALID_ENUM"},
		{name: "invalid json", code: schema.SchemaIssueCodeRuleInvalidJSON, expected: "RULE_INVALID_JSON"},
		{name: "invalid text", code: schema.SchemaIssueCodeRuleInvalidText, expected: "RULE_INVALID_TEXT"},
		{name: "sensitive value", code: schema.SchemaIssueCodeRuleSensitiveValueNotAllowed, expected: "RULE_SENSITIVE_VALUE_NOT_ALLOWED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.code.IsKnown() {
				t.Fatalf("%s should extend the known SchemaIssueCode set", tt.code)
			}
			if tt.code.IsUnknown() {
				t.Fatalf("%s should not be unknown", tt.code)
			}
			if got := tt.code.String(); got != tt.expected {
				t.Fatalf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestValidateGeneratorConfigAcceptsValidDraftAndPersistedModels(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	draft := validGeneratorConfig()
	draft.ID = 0
	draft.CreatedAt = time.Time{}
	draft.UpdatedAt = time.Time{}
	if issues := ValidateGeneratorConfig(draft, schema.SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateGeneratorConfig(valid draft) = %#v, want no issues", issues)
	}

	persisted := validGeneratorConfig()
	persisted.ID = 11
	persisted.CreatedAt = createdAt
	persisted.UpdatedAt = updatedAt
	if issues := ValidateGeneratorConfig(persisted, schema.SchemaValidationModePersisted); len(issues) != 0 {
		t.Fatalf("ValidateGeneratorConfig(valid persisted) = %#v, want no issues", issues)
	}
}

func TestValidateGeneratorConfigReturnsMultipleDomainOnlyIssues(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	config := GeneratorConfig{
		ID:              -1,
		ColumnID:        0,
		GeneratorName:   "bad/name",
		DataMappingType: DataMappingType("json"),
		Params:          GeneratorParams{Raw: json.RawMessage(`{"apiKey":"value"}`)},
		ConfigStatus:    ConfigStatus("DISABLED"),
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt.Add(-time.Minute),
	}

	issues := ValidateGeneratorConfig(config, schema.SchemaValidationModeDraft)
	assertRuleIssueCodesByPath(t, issues, map[string]schema.SchemaIssueCode{
		"id":              schema.SchemaIssueCodeInvalidID,
		"columnId":        schema.SchemaIssueCodeInvalidID,
		"generatorName":   schema.SchemaIssueCodeRuleInvalidText,
		"dataMappingType": schema.SchemaIssueCodeRuleInvalidEnum,
		"params":          schema.SchemaIssueCodeRuleSensitiveValueNotAllowed,
		"configStatus":    schema.SchemaIssueCodeRuleInvalidEnum,
		"updatedAt":       schema.SchemaIssueCodeInvalidTime,
	})
}

func TestValidateGeneratorConfigUsesPersistedModeForIDAndAuditTimes(t *testing.T) {
	config := validGeneratorConfig()
	config.ID = 0
	config.CreatedAt = time.Time{}
	config.UpdatedAt = time.Time{}

	issues := ValidateGeneratorConfig(config, schema.SchemaValidationModePersisted)
	assertRuleIssueCodesByPath(t, issues, map[string]schema.SchemaIssueCode{
		"id":        schema.SchemaIssueCodeInvalidID,
		"createdAt": schema.SchemaIssueCodeInvalidTime,
		"updatedAt": schema.SchemaIssueCodeInvalidTime,
	})
}

func TestValidateGeneratorConfigRejectsUnknownMode(t *testing.T) {
	issues := ValidateGeneratorConfig(validGeneratorConfig(), schema.SchemaValidationMode("runtime"))
	assertRuleIssueCodesByPath(t, issues, map[string]schema.SchemaIssueCode{
		"mode": schema.SchemaIssueCodeValidationFailed,
	})
}

func TestValidateGeneratorConfigRejectsRequiredFieldOmissions(t *testing.T) {
	config := validGeneratorConfig()
	config.GeneratorName = " \t "
	config.DataMappingType = ""
	config.ConfigStatus = ""

	issues := ValidateGeneratorConfig(config, schema.SchemaValidationModeDraft)
	assertRuleIssueCodesByPath(t, issues, map[string]schema.SchemaIssueCode{
		"generatorName":   schema.SchemaIssueCodeRequired,
		"dataMappingType": schema.SchemaIssueCodeRequired,
		"configStatus":    schema.SchemaIssueCodeRequired,
	})
}

func TestValidateGeneratorConfigRejectsGeneratorNameBoundaryValues(t *testing.T) {
	tests := []struct {
		name          string
		generatorName string
	}{
		{name: "too long", generatorName: strings.Repeat("a", 101)},
		{name: "forward slash", generatorName: "faker/person/name"},
		{name: "backslash", generatorName: `faker\person\name`},
		{name: "control character", generatorName: "faker.person.name\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validGeneratorConfig()
			config.GeneratorName = tt.generatorName

			issues := ValidateGeneratorConfig(config, schema.SchemaValidationModeDraft)
			assertRuleIssueCodesByPath(t, issues, map[string]schema.SchemaIssueCode{
				"generatorName": schema.SchemaIssueCodeRuleInvalidText,
			})
		})
	}
}

func TestValidateGeneratorConfigRejectsUnsafeGeneratorNameControlCharacters(t *testing.T) {
	config := validGeneratorConfig()
	config.GeneratorName = "faker.person.name\n"

	issues := ValidateGeneratorConfig(config, schema.SchemaValidationModeDraft)
	assertRuleIssueCodesByPath(t, issues, map[string]schema.SchemaIssueCode{
		"generatorName": schema.SchemaIssueCodeRuleInvalidText,
	})
}

func TestValidateGeneratorParamsRejectsInvalidJSONAndSensitiveFieldNames(t *testing.T) {
	tests := []struct {
		name   string
		params GeneratorParams
		code   schema.SchemaIssueCode
	}{
		{name: "invalid json", params: GeneratorParams{Raw: json.RawMessage(`{"bad":`)}, code: schema.SchemaIssueCodeRuleInvalidJSON},
		{name: "password key", params: GeneratorParams{Raw: json.RawMessage(`{"password":"value"}`)}, code: schema.SchemaIssueCodeRuleSensitiveValueNotAllowed},
		{name: "nested token key", params: GeneratorParams{Raw: json.RawMessage(`{"nested":{"refreshToken":"value"}}`)}, code: schema.SchemaIssueCodeRuleSensitiveValueNotAllowed},
		{name: "array secret key", params: GeneratorParams{Raw: json.RawMessage(`[{"client_secret":"value"}]`)}, code: schema.SchemaIssueCodeRuleSensitiveValueNotAllowed},
		{name: "hyphenated api key", params: GeneratorParams{Raw: json.RawMessage(`{"api-key":"value"}`)}, code: schema.SchemaIssueCodeRuleSensitiveValueNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := ValidateGeneratorParams(tt.params)
			assertRuleIssueCodesByPath(t, issues, map[string]schema.SchemaIssueCode{"params": tt.code})
		})
	}

	for _, params := range []GeneratorParams{{}, {Raw: json.RawMessage("null")}, {Raw: json.RawMessage(`{"locale":"zh-CN","enabled":true}`)}, {Raw: json.RawMessage(`[1,"safe",true]`)}} {
		if issues := ValidateGeneratorParams(params); len(issues) != 0 {
			t.Fatalf("ValidateGeneratorParams(%s) = %#v, want no issues", params.Raw, issues)
		}
	}
}

func TestValidateGeneratorParamsDoesNotValidateGeneratorSpecificSchema(t *testing.T) {
	params := GeneratorParams{Raw: json.RawMessage(`{"min":99,"max":1,"enum":[],"unsupportedOption":true}`)}

	if issues := ValidateGeneratorParams(params); len(issues) != 0 {
		t.Fatalf("ValidateGeneratorParams(generator-specific schema boundary) = %#v, want no domain-only issues", issues)
	}
}

func TestValidateGeneratorConfigDoesNotValidateDeferredExternalState(t *testing.T) {
	config := validGeneratorConfig()
	config.ColumnID = 999999
	config.GeneratorName = "future.registry.generator"
	config.Params = GeneratorParams{Raw: json.RawMessage(`{"unsupportedGeneratorSpecificOption":true}`)}

	if issues := ValidateGeneratorConfig(config, schema.SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateGeneratorConfig(deferred external state) = %#v, want no domain-only issues", issues)
	}
}

func TestRulePackageDoesNotDeclareParallelValidationContracts(t *testing.T) {
	for _, file := range ruleProductionFiles(t) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatalf("parse rule file %s: %v", file, err)
		}

		for _, decl := range parsed.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				name := typeSpec.Name.Name
				for _, forbidden := range []string{"RuleValidationIssue", "RuleIssueCode", "RuleIssueSeverity", "RuleValidationMode"} {
					if name == forbidden {
						t.Fatalf("rule package must not declare parallel validation contract type %s in %s", name, file)
					}
				}
			}
		}
	}
}

func TestRulePackageProductionFilesOnlyImportAllowedValidationDependencies(t *testing.T) {
	for _, file := range ruleProductionFiles(t) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse imports for %s: %v", file, err)
		}

		for _, importSpec := range parsed.Imports {
			importPath := strings.Trim(importSpec.Path.Value, "\"")
			for _, forbidden := range []string{
				"github.com/gerdong/loomidbx/internal/config",
				"github.com/gerdong/loomidbx/internal/service",
				"github.com/gerdong/loomidbx/internal/repository",
				"github.com/gerdong/loomidbx/internal/storage",
				"github.com/gerdong/loomidbx/internal/dbx",
				"github.com/wailsapp/wails",
				"database/sql",
			} {
				if importPath == forbidden || strings.HasPrefix(importPath, forbidden+"/") {
					t.Fatalf("rule domain file %s imports out-of-scope package %q", file, importPath)
				}
			}
		}
	}
}

func TestRulePackageProductionFilesDoNotUseOutOfScopeValidationAPIs(t *testing.T) {
	for _, file := range ruleProductionFiles(t) {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read rule package file %s: %v", file, err)
		}
		text := string(content)
		for _, forbidden := range []string{"sql.", "db.", "repository.", "service.", "wails.", "runtime.", "exec.", "http.", "registry."} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("rule domain file %s uses out-of-scope validation API marker %q", file, forbidden)
			}
		}
	}
}

func assertRuleValidationIssue(t *testing.T, got schema.SchemaValidationIssue, want schema.SchemaValidationIssue) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("issue = %#v, want %#v", got, want)
	}
	if contractIssues := schema.ValidateIssue(got); len(contractIssues) != 0 {
		t.Fatalf("issue does not satisfy schema validation contract: %#v", contractIssues)
	}
	message := strings.ToLower(got.Message)
	for _, forbidden := range []string{"password", "secret", "token", "apikey", "api_key"} {
		if strings.Contains(message, forbidden) {
			t.Fatalf("issue message should stay safe and generic: %#v", got)
		}
	}
}

func assertRuleIssueCodesByPath(t *testing.T, issues []schema.SchemaValidationIssue, want map[string]schema.SchemaIssueCode) {
	t.Helper()

	if got := len(issues); got != len(want) {
		t.Fatalf("issues count = %d, want %d: %#v", got, len(want), issues)
	}
	for _, issue := range issues {
		wantCode, ok := want[issue.Path]
		if !ok {
			t.Fatalf("unexpected issue path %q in %#v", issue.Path, issues)
		}
		assertRuleValidationIssue(t, issue, schema.SchemaValidationIssue{
			Path:     issue.Path,
			Code:     wantCode,
			Severity: schema.SchemaIssueSeverityError,
			Message:  issue.Message,
		})
	}
}

func validGeneratorConfig() GeneratorConfig {
	createdAt := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	return GeneratorConfig{
		ID:              1,
		ColumnID:        42,
		GeneratorName:   "faker.person.name",
		DataMappingType: DataMappingTypeText,
		Params:          GeneratorParams{Raw: json.RawMessage(`{"locale":"en","enabled":true}`)},
		ConfigStatus:    ConfigStatusActive,
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt.Add(time.Minute),
	}
}

func ruleProductionFiles(t *testing.T) []string {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(".", "*.go"))
	if err != nil {
		t.Fatalf("glob rule package files: %v", err)
	}

	productionFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read rule package file %s: %v", file, err)
		}
		if len(strings.TrimSpace(string(content))) == 0 {
			t.Fatalf("rule package production file %s must not be empty", file)
		}
		productionFiles = append(productionFiles, file)
	}
	return productionFiles
}
