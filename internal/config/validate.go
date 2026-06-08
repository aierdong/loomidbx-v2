package config

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
)

var sensitivePlaintextPattern = regexp.MustCompile(`(?i)(api[_-]?key|token|password|passwd|secret|connection[_-]?string|bearer\s+|sk-[a-z0-9])`)

// ConfigValidator validates ordinary application configuration for load and save flows.
type ConfigValidator struct{}

// ValidateAppConfig validates a fully merged AppConfig.
//
// The returned issues are field-level, user-readable, and sanitized so callers can reject invalid load
// results without continuing to use them as valid configuration.
func (ConfigValidator) ValidateAppConfig(cfg AppConfig) []ConfigIssue {
	var issues []ConfigIssue
	issues = append(issues, validateLanguage(cfg.Appearance.Language)...)
	issues = append(issues, validateTheme(cfg.Appearance.Theme)...)
	issues = append(issues, validateRunMode(cfg.Development.Mode)...)
	issues = append(issues, validateAbsolutePath("paths.dataDir", cfg.Paths.DataDir)...)
	issues = append(issues, validateAbsolutePath("paths.configFile", cfg.Paths.ConfigFile)...)
	issues = append(issues, validateFutureIntegration("integrations.account", cfg.Integrations.Account)...)
	issues = append(issues, validateFutureIntegration("integrations.llm", cfg.Integrations.LLM)...)
	issues = append(issues, validatePrivacyBoundary(cfg.Privacy)...)
	return issues
}

// ValidateUserConfig validates the persisted user-editable config shape before saving.
//
// It reuses the same field rules as ValidateAppConfig where fields are present and also scans the
// persisted JSON representation for sensitive plaintext markers that ordinary config must reject.
func (validator ConfigValidator) ValidateUserConfig(cfg UserConfig) []ConfigIssue {
	var issues []ConfigIssue
	if cfg.Appearance != nil {
		if cfg.Appearance.Language != "" {
			issues = append(issues, validateLanguage(cfg.Appearance.Language)...)
		}
		if cfg.Appearance.Theme != "" {
			issues = append(issues, validateTheme(cfg.Appearance.Theme)...)
		}
	}
	if cfg.Paths != nil && cfg.Paths.DataDir != "" {
		issues = append(issues, validateAbsolutePath("paths.dataDir", cfg.Paths.DataDir)...)
	}
	if cfg.Development != nil {
		if cfg.Development.Mode != "" {
			issues = append(issues, validateRunMode(cfg.Development.Mode)...)
		}
	}
	if cfg.Integrations != nil {
		issues = append(issues, validateFutureIntegration("integrations.account", cfg.Integrations.Account)...)
		issues = append(issues, validateFutureIntegration("integrations.llm", cfg.Integrations.LLM)...)
	}

	if raw, err := json.Marshal(cfg); err == nil {
		issues = append(issues, validator.ValidateConfigJSON(raw)...)
	}
	return issues
}

// ValidateConfigJSON scans raw ordinary config JSON for sensitive plaintext fields or values.
//
// Invalid JSON is ignored here because parsing failures are owned by ConfigFileStore; this method only
// enforces the privacy boundary and avoids echoing raw sensitive values in issue messages.
func (ConfigValidator) ValidateConfigJSON(raw []byte) []ConfigIssue {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	return scanSensitiveJSON("", value)
}

// ValidateForLoad validates a merged AppConfig using the shared ConfigValidator rules.
func (validator ConfigValidator) ValidateForLoad(cfg AppConfig) []ConfigIssue {
	return validator.ValidateAppConfig(cfg)
}

// ValidateForSave validates a UserConfig using the shared ConfigValidator rules.
func (validator ConfigValidator) ValidateForSave(cfg UserConfig) []ConfigIssue {
	return validator.ValidateUserConfig(cfg)
}

func validateLanguage(value Language) []ConfigIssue {
	switch value {
	case LanguageZh, LanguageEn:
		return nil
	default:
		return []ConfigIssue{validationIssue("appearance.language", "语言必须是 zh 或 en")}
	}
}

func validateTheme(value Theme) []ConfigIssue {
	switch value {
	case ThemeLight, ThemeDark, ThemeSystem:
		return nil
	default:
		return []ConfigIssue{validationIssue("appearance.theme", "主题必须是 light、dark 或 system")}
	}
}

func validateRunMode(value RunMode) []ConfigIssue {
	switch value {
	case ModeDesktop, ModeDevelopment, ModeTest:
		return nil
	default:
		return []ConfigIssue{validationIssue("development.mode", "运行模式必须是 desktop、development 或 test")}
	}
}

func validateAbsolutePath(path string, value string) []ConfigIssue {
	if value == "" {
		return []ConfigIssue{invalidPathIssue(path, "路径不能为空")}
	}
	if !filepath.IsAbs(value) {
		return []ConfigIssue{invalidPathIssue(path, "路径必须是绝对路径")}
	}
	return nil
}

func validateFutureIntegration(prefix string, value FutureIntegrationConfig) []ConfigIssue {
	var issues []ConfigIssue
	switch value.Status {
	case FutureStatusUnavailable, FutureStatusNotConfigured, FutureStatusConfigured:
	default:
		issues = append(issues, validationIssue(prefix+".status", "未来入口状态必须是受支持的枚举值"))
	}

	if value.Enabled && value.Status == FutureStatusUnavailable {
		issues = append(issues, validationIssue(prefix+".enabled", "未来入口未完成时不能标记为启用"))
	}
	if value.Status == FutureStatusConfigured && !value.Configured {
		issues = append(issues, validationIssue(prefix+".status", "状态与已配置标记不一致"))
	}
	if value.Configured && value.Status != FutureStatusConfigured {
		issues = append(issues, validationIssue(prefix+".status", "状态与已配置标记不一致"))
	}
	return issues
}

func validatePrivacyBoundary(value PrivacyConfig) []ConfigIssue {
	var issues []ConfigIssue
	if value.BusinessData.StoredInAppConfig {
		issues = append(issues, validationIssue("privacy.businessData.storedInAppConfig", "SQLite 业务存储属于 phase-01-local-storage-strategy 相邻 spec，不能存入普通应用配置"))
	}
	if value.SensitiveCredentials.StoredInAppConfig {
		issues = append(issues, sensitiveValueIssue("privacy.sensitiveCredentials.storedInAppConfig"))
	}
	if value.SensitiveCredentials.PlaintextPolicy != "" && sensitivePlaintextPattern.MatchString(value.SensitiveCredentials.PlaintextPolicy) {
		issues = append(issues, sensitiveValueIssue("privacy.sensitiveCredentials.plaintextPolicy"))
	}
	return issues
}

func scanSensitiveJSON(prefix string, value any) []ConfigIssue {
	switch typed := value.(type) {
	case map[string]any:
		var issues []ConfigIssue
		for key, child := range typed {
			path := joinIssuePath(prefix, key)
			if sensitivePlaintextPattern.MatchString(key) {
				issues = append(issues, sensitiveValueIssue(path))
				continue
			}
			issues = append(issues, scanSensitiveJSON(path, child)...)
		}
		return issues
	case []any:
		var issues []ConfigIssue
		for _, child := range typed {
			issues = append(issues, scanSensitiveJSON(prefix, child)...)
		}
		return issues
	case string:
		if sensitivePlaintextPattern.MatchString(typed) {
			return []ConfigIssue{sensitiveValueIssue(prefix)}
		}
	}
	return nil
}

func joinIssuePath(prefix string, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func validationIssue(path string, message string) ConfigIssue {
	return ConfigIssue{
		Path:     path,
		Code:     ConfigIssueCodeValidationFailed,
		Severity: ConfigIssueSeverityError,
		Message:  message,
	}
}

func sensitiveValueIssue(path string) ConfigIssue {
	path = strings.Trim(path, ".")
	if path == "" {
		path = "config"
	}
	return ConfigIssue{
		Path:     path,
		Code:     ConfigIssueCodeSensitiveValueNotAllowed,
		Severity: ConfigIssueSeverityError,
		Message:  "普通配置不允许保存敏感明文，请使用后续安全存储边界",
	}
}
