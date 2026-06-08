package config

import (
	"path/filepath"
	"testing"
)

func TestConfigValidatorAcceptsValidDefaultConfig(t *testing.T) {
	cfg := validTestAppConfig(t)

	issues := ConfigValidator{}.ValidateAppConfig(cfg)

	assertNoValidationIssues(t, issues)
}

func TestConfigValidatorReportsInvalidLanguageThemeAndMode(t *testing.T) {
	cfg := validTestAppConfig(t)
	cfg.Appearance.Language = Language("fr")
	cfg.Appearance.Theme = Theme("blue")
	cfg.Development.Mode = RunMode("server")

	issues := ConfigValidator{}.ValidateAppConfig(cfg)

	assertIssue(t, issues, "appearance.language", ConfigIssueCodeValidationFailed)
	assertIssue(t, issues, "appearance.theme", ConfigIssueCodeValidationFailed)
	assertIssue(t, issues, "development.mode", ConfigIssueCodeValidationFailed)
}

func TestConfigValidatorReportsInvalidFutureIntegrationStatus(t *testing.T) {
	cfg := validTestAppConfig(t)
	cfg.Integrations.Account.Status = FutureIntegrationStatus("ready")
	cfg.Integrations.LLM.Status = FutureIntegrationStatus("connected")

	issues := ConfigValidator{}.ValidateAppConfig(cfg)

	assertIssue(t, issues, "integrations.account.status", ConfigIssueCodeValidationFailed)
	assertIssue(t, issues, "integrations.llm.status", ConfigIssueCodeValidationFailed)
}

func TestConfigValidatorReportsRelativeAndEmptyPaths(t *testing.T) {
	cfg := validTestAppConfig(t)
	cfg.Paths.DataDir = "relative-data"
	cfg.Paths.ConfigFile = ""

	issues := ConfigValidator{}.ValidateAppConfig(cfg)

	assertIssue(t, issues, "paths.dataDir", ConfigIssueCodeConfigPathInvalid)
	assertIssue(t, issues, "paths.configFile", ConfigIssueCodeConfigPathInvalid)
}

func TestConfigValidatorReportsFutureIntegrationStateInconsistency(t *testing.T) {
	cfg := validTestAppConfig(t)
	cfg.Integrations.Account = FutureIntegrationConfig{
		Enabled:    true,
		Configured: false,
		Status:     FutureStatusUnavailable,
	}
	cfg.Integrations.LLM = FutureIntegrationConfig{
		Enabled:    false,
		Configured: true,
		Status:     FutureStatusNotConfigured,
	}

	issues := ConfigValidator{}.ValidateAppConfig(cfg)

	assertIssue(t, issues, "integrations.account.enabled", ConfigIssueCodeValidationFailed)
	assertIssue(t, issues, "integrations.llm.status", ConfigIssueCodeValidationFailed)
}

func TestConfigValidatorAllowsConfiguredFutureStatusMarker(t *testing.T) {
	cfg := validTestAppConfig(t)
	cfg.Integrations.LLM = FutureIntegrationConfig{
		Enabled:    false,
		Configured: true,
		Status:     FutureStatusConfigured,
	}

	issues := ConfigValidator{}.ValidateAppConfig(cfg)

	assertNoValidationIssues(t, issues)
}

func TestConfigValidatorRejectsSensitivePlaintextInConfigJSONWithoutLeakingValue(t *testing.T) {
	rawValue := "bearer credential"
	keyName := "api" + "Key"
	raw := []byte(`{
		"version": 1,
		"appearance": {"language": "zh", "theme": "system"},
		"integrations": {"llm": {"` + keyName + `": "` + rawValue + `"}}
	}`)

	issues := ConfigValidator{}.ValidateConfigJSON(raw)

	assertIssue(t, issues, "integrations.llm."+keyName, ConfigIssueCodeSensitiveValueNotAllowed)
	assertMessagesDoNotContain(t, issues, rawValue)
}

func TestConfigValidatorRejectsSensitivePlaintextMarkerInUserConfig(t *testing.T) {
	rawValue := "bearer credential"
	user := UserConfig{
		Version: CurrentConfigVersion,
		Paths:   &UserPathConfig{},
	}
	raw := []byte(`{"paths":{"dataDir":"` + jsonPath(filepath.Join(t.TempDir(), "data")) + `"},"integrations":{"llm":{"credential":"` + rawValue + `"}}}`)

	issues := append(ConfigValidator{}.ValidateUserConfig(user), ConfigValidator{}.ValidateConfigJSON(raw)...)

	assertIssue(t, issues, "integrations.llm.credential", ConfigIssueCodeSensitiveValueNotAllowed)
	assertMessagesDoNotContain(t, issues, rawValue)
}

func TestConfigValidatorProvidesReusableLoadAndSaveAPIs(t *testing.T) {
	cfg := validTestAppConfig(t)
	user := DefaultUserConfig()
	user.Paths.DataDir = cfg.Paths.DataDir

	loadIssues := ConfigValidator{}.ValidateForLoad(cfg)
	saveIssues := ConfigValidator{}.ValidateForSave(user)

	assertNoValidationIssues(t, loadIssues)
	assertNoValidationIssues(t, saveIssues)
}

func validTestAppConfig(t *testing.T) AppConfig {
	t.Helper()

	root := t.TempDir()
	cfg := DefaultAppConfig()
	cfg.Paths.ConfigFile = filepath.Join(root, "config", "config.json")
	cfg.Paths.DataDir = filepath.Join(root, "data")
	return cfg
}

func assertNoValidationIssues(t *testing.T, issues []ConfigIssue) {
	t.Helper()
	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}
