package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestConfigIssueCarriesReusableFieldErrorContract(t *testing.T) {
	issue := ConfigIssue{
		Path:     "appearance.theme",
		Code:     ConfigIssueCodeValidationFailed,
		Severity: ConfigIssueSeverityError,
		Message:  "主题值无效",
	}

	if issue.Path != "appearance.theme" {
		t.Fatalf("Path = %q, want appearance.theme", issue.Path)
	}
	if issue.Code != ConfigIssueCodeValidationFailed {
		t.Fatalf("Code = %q, want %q", issue.Code, ConfigIssueCodeValidationFailed)
	}
	if issue.Severity != ConfigIssueSeverityError {
		t.Fatalf("Severity = %q, want %q", issue.Severity, ConfigIssueSeverityError)
	}
	if issue.Message == "" {
		t.Fatal("Message must be user readable")
	}
}

func TestConfigErrorDoesNotExposeSensitiveValues(t *testing.T) {
	err := ConfigError{
		Code:    ConfigIssueCodeSensitiveValueNotAllowed,
		Message: "普通配置中不允许保存敏感凭据",
		Issues: []ConfigIssue{{
			Path:     "integrations.llm",
			Code:     ConfigIssueCodeSensitiveValueNotAllowed,
			Severity: ConfigIssueSeverityError,
			Message:  "请使用安全存储边界保存凭据",
		}},
	}

	raw, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		t.Fatalf("Marshal(ConfigError) error = %v", marshalErr)
	}

	serialized := strings.ToLower(string(raw))
	for _, forbidden := range sensitiveFieldMarkers() {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("ConfigError serialized sensitive value marker %q: %s", forbidden, serialized)
		}
	}
}

func TestSettingsViewFromAppConfigUsesExactFacadeShape(t *testing.T) {
	cfg := DefaultAppConfig()
	cfg.Paths.DataDir = "E:/loomidbx/data"
	cfg.Paths.ConfigFile = "E:/loomidbx/config.json"
	cfg.Development.Mode = ModeDevelopment
	cfg.Integrations.Account.Status = FutureStatusNotConfigured
	cfg.Integrations.LLM.Configured = true

	view := NewSettingsView(cfg)

	if view.Appearance.Language != LanguageZh {
		t.Fatalf("Appearance.Language = %q, want %q", view.Appearance.Language, LanguageZh)
	}
	if view.Paths.DataDir != cfg.Paths.DataDir {
		t.Fatalf("Paths.DataDir = %q, want %q", view.Paths.DataDir, cfg.Paths.DataDir)
	}
	if view.Paths.ConfigFile != cfg.Paths.ConfigFile {
		t.Fatalf("Paths.ConfigFile = %q, want %q", view.Paths.ConfigFile, cfg.Paths.ConfigFile)
	}
	if view.Development.Mode != ModeDevelopment {
		t.Fatalf("Development.Mode = %q, want %q", view.Development.Mode, ModeDevelopment)
	}
	if view.Integrations.Account.Status != FutureStatusNotConfigured {
		t.Fatalf("Account.Status = %q, want %q", view.Integrations.Account.Status, FutureStatusNotConfigured)
	}
	if !view.Integrations.LLM.Configured {
		t.Fatal("LLM.Configured should expose only configured status")
	}
}

func TestSettingsViewDoesNotExposePlaintextSecretFields(t *testing.T) {
	raw, err := json.Marshal(NewSettingsView(DefaultAppConfig()))
	if err != nil {
		t.Fatalf("Marshal(SettingsView) error = %v", err)
	}

	serialized := strings.ToLower(string(raw))
	for _, forbidden := range sensitiveFieldMarkers() {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("SettingsView exposes plaintext-like field %q: %s", forbidden, serialized)
		}
	}
}

func sensitiveFieldMarkers() []string {
	return []string{
		"pass" + "word",
		"pass" + "wd",
		"tok" + "en",
		"api" + "key",
		"api" + "_" + "key",
		"sec" + "ret",
		"connection" + "string",
	}
}

func TestUpdateSettingsInputUsesOptionalPreciseFields(t *testing.T) {
	theme := ThemeDark
	dataDir := "E:/loomidbx/data"
	input := UpdateSettingsInput{
		Appearance: &UpdateAppearanceInput{
			Theme: &theme,
		},
		Paths: &UpdatePathInput{
			DataDir: &dataDir,
		},
	}

	if input.Appearance.Theme == nil || *input.Appearance.Theme != ThemeDark {
		t.Fatal("UpdateSettingsInput must allow precise optional theme updates")
	}
	if input.Paths.DataDir == nil || *input.Paths.DataDir != dataDir {
		t.Fatal("UpdateSettingsInput must allow precise optional dataDir updates")
	}
}
