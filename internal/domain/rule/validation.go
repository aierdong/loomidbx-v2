package rule

import (
	"encoding/json"
	"strings"
	"time"
	"unicode"

	"github.com/gerdong/loomidbx/internal/domain/schema"
)

// ValidateGeneratorConfig validates a generator config using only rule-domain invariants.
func ValidateGeneratorConfig(config GeneratorConfig, mode schema.SchemaValidationMode) []schema.SchemaValidationIssue {
	issues := make([]schema.SchemaValidationIssue, 0, 8)
	if mode.IsUnknown() {
		return append(issues, ruleValidationIssue("mode", schema.SchemaIssueCodeValidationFailed, "mode must be draft or persisted"))
	}

	issues = append(issues, validateGeneratorConfigID(config.ID, mode)...)
	if config.ColumnID <= 0 {
		issues = append(issues, ruleValidationIssue("columnId", schema.SchemaIssueCodeInvalidID, "columnId must reference a valid column identity"))
	}
	issues = append(issues, validateGeneratorName(config.GeneratorName)...)
	if config.DataMappingType == "" {
		issues = append(issues, ruleValidationIssue("dataMappingType", schema.SchemaIssueCodeRequired, "dataMappingType is required"))
	} else if config.DataMappingType.IsUnknown() {
		issues = append(issues, ruleValidationIssue("dataMappingType", schema.SchemaIssueCodeRuleInvalidEnum, "dataMappingType must be one of the stable mapping values"))
	}
	if config.ConfigStatus == "" {
		issues = append(issues, ruleValidationIssue("configStatus", schema.SchemaIssueCodeRequired, "configStatus is required"))
	} else if config.ConfigStatus.IsUnknown() {
		issues = append(issues, ruleValidationIssue("configStatus", schema.SchemaIssueCodeRuleInvalidEnum, "configStatus must be one of the stable status values"))
	}
	issues = append(issues, ValidateGeneratorParams(config.Params)...)
	issues = append(issues, validateGeneratorConfigAuditTimes(mode, config.CreatedAt, config.UpdatedAt)...)
	return issues
}

// ValidateGeneratorParams validates the rule params JSON payload without binding to a generator schema.
func ValidateGeneratorParams(params GeneratorParams) []schema.SchemaValidationIssue {
	if len(params.Raw) == 0 || string(params.Raw) == "null" {
		return nil
	}

	var value any
	if err := json.Unmarshal(params.Raw, &value); err != nil {
		return []schema.SchemaValidationIssue{ruleValidationIssue("params", schema.SchemaIssueCodeRuleInvalidJSON, "params must be valid JSON")}
	}
	if containsSensitiveGeneratorParamKey(value) {
		return []schema.SchemaValidationIssue{ruleValidationIssue("params", schema.SchemaIssueCodeRuleSensitiveValueNotAllowed, "params must not contain credential field names")}
	}
	return nil
}

func ruleValidationIssue(path string, code schema.SchemaIssueCode, message string) schema.SchemaValidationIssue {
	return schema.SchemaValidationIssue{
		Path:     path,
		Code:     code,
		Severity: schema.SchemaIssueSeverityError,
		Message:  message,
	}
}

func validateGeneratorConfigID(id int64, mode schema.SchemaValidationMode) []schema.SchemaValidationIssue {
	if id < 0 || mode == schema.SchemaValidationModePersisted && id <= 0 {
		return []schema.SchemaValidationIssue{ruleValidationIssue("id", schema.SchemaIssueCodeInvalidID, "id must satisfy the validation mode identity rule")}
	}
	return nil
}

func validateGeneratorName(name string) []schema.SchemaValidationIssue {
	if strings.TrimSpace(name) == "" {
		return []schema.SchemaValidationIssue{ruleValidationIssue("generatorName", schema.SchemaIssueCodeRequired, "generatorName is required")}
	}
	if len([]rune(name)) > 100 || containsUnsafeGeneratorNameRune(name) {
		return []schema.SchemaValidationIssue{ruleValidationIssue("generatorName", schema.SchemaIssueCodeRuleInvalidText, "generatorName must be a safe generator identifier")}
	}
	return nil
}

func containsUnsafeGeneratorNameRune(name string) bool {
	for _, r := range name {
		if unicode.IsControl(r) || r == '/' || r == '\\' {
			return true
		}
	}
	return false
}

func validateGeneratorConfigAuditTimes(mode schema.SchemaValidationMode, createdAt time.Time, updatedAt time.Time) []schema.SchemaValidationIssue {
	issues := make([]schema.SchemaValidationIssue, 0, 2)
	if mode == schema.SchemaValidationModePersisted {
		if createdAt.IsZero() {
			issues = append(issues, ruleValidationIssue("createdAt", schema.SchemaIssueCodeInvalidTime, "createdAt is required for persisted snapshots"))
		}
		if updatedAt.IsZero() {
			issues = append(issues, ruleValidationIssue("updatedAt", schema.SchemaIssueCodeInvalidTime, "updatedAt is required for persisted snapshots"))
		}
	}
	if !createdAt.IsZero() && !updatedAt.IsZero() && updatedAt.Before(createdAt) {
		issues = append(issues, ruleValidationIssue("updatedAt", schema.SchemaIssueCodeInvalidTime, "updatedAt must not be earlier than createdAt"))
	}
	return issues
}

func containsSensitiveGeneratorParamKey(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			if isSensitiveGeneratorParamKey(key) || containsSensitiveGeneratorParamKey(nested) {
				return true
			}
		}
	case []any:
		for _, nested := range typed {
			if containsSensitiveGeneratorParamKey(nested) {
				return true
			}
		}
	}
	return false
}

func isSensitiveGeneratorParamKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
	for _, sensitive := range []string{"password", "token", "secret", "apikey"} {
		if strings.Contains(normalized, sensitive) {
			return true
		}
	}
	return false
}
