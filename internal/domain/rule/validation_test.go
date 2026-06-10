package rule

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

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
