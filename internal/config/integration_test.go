package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigOutputFeedsLocalStorageStrategyProbeWithoutSQLiteSchema(t *testing.T) {
	root := t.TempDir()
	service := newTestConfigService(root, JSONConfigFileStore{}, map[string]string{
		EnvMode:     string(ModeTest),
		EnvTestRoot: filepath.Join(root, "isolated-test-root"),
	})

	view, err := service.Current()

	if err != nil {
		t.Fatalf("Current() error = %v, want nil", err)
	}
	probe := localStorageStrategyProbe{}
	storageConfig, err := probe.ConfigureFromSettings(view)
	if err != nil {
		t.Fatalf("ConfigureFromSettings() error = %v, want nil", err)
	}
	assertPathEqual(t, storageConfig.DataDir, filepath.Join(root, "isolated-test-root", "data"))
	if storageConfig.Mode != ModeTest {
		t.Fatalf("Mode = %q, want %q", storageConfig.Mode, ModeTest)
	}
	if probe.schemaCreated || probe.migrationsRun {
		t.Fatalf("probe touched SQLite schema or migrations: %+v", probe)
	}

	readme := readConfigReadme(t)
	for _, required := range []string{"SQLite", "phase-01-local-storage-strategy", "相邻 spec"} {
		if !strings.Contains(readme, required) {
			t.Fatalf("internal/config/README.md missing %q boundary note:\n%s", required, readme)
		}
	}

	issues := ConfigValidator{}.ValidateAppConfig(AppConfig{
		Version: CurrentConfigVersion,
		Appearance: AppearanceConfig{
			Language: LanguageZh,
			Theme:    ThemeSystem,
		},
		Paths: PathConfig{
			DataDir:    view.Paths.DataDir,
			ConfigFile: view.Paths.ConfigFile,
		},
		Development: DevelopmentConfig{
			Mode: ModeTest,
		},
		Integrations: DefaultAppConfig().Integrations,
		Privacy: PrivacyConfig{
			LocalOnly: true,
			BusinessData: BusinessDataBoundary{
				StoredInAppConfig: true,
			},
			SensitiveCredentials: SensitiveCredentialsBoundary{
				ExternalStorageRequired: true,
			},
			NetworkUploadDisabled: true,
		},
	})
	assertIssue(t, issues, "privacy.businessData.storedInAppConfig", ConfigIssueCodeValidationFailed)
	serializedIssues := joinIssueMessages(issues)
	for _, required := range []string{"SQLite", "phase-01-local-storage-strategy", "相邻 spec"} {
		if !strings.Contains(serializedIssues, required) {
			t.Fatalf("business storage issue message missing %q boundary note: %+v", required, issues)
		}
	}
}

func TestConfigServiceRejectsSensitivePlaintextInOrdinaryConfigWithoutLeakingValue(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config-root", "LoomiDBX", "config.json")
	secretValue := "bearer " + "load-secret"
	sensitiveKey := "api" + "Key"
	writeRawFile(t, configPath, `{
		"version": 1,
		"appearance": {"language": "zh", "theme": "system"},
		"integrations": {"llm": {"`+sensitiveKey+`": "`+secretValue+`"}}
	}`)
	service := newTestConfigService(root, JSONConfigFileStore{}, nil)

	_, err := service.Load()

	configErr, ok := err.(ConfigError)
	if !ok {
		t.Fatalf("Load() error = %T %[1]v, want ConfigError", err)
	}
	if configErr.Code != ConfigIssueCodeSensitiveValueNotAllowed {
		t.Fatalf("Code = %q, want %q", configErr.Code, ConfigIssueCodeSensitiveValueNotAllowed)
	}
	assertIssue(t, configErr.Issues, "integrations.llm."+sensitiveKey, ConfigIssueCodeSensitiveValueNotAllowed)
	assertConfigErrorDoesNotContain(t, configErr, secretValue, "load-secret", sensitiveKey)
}

func TestConfigServiceSaveReloadRoundTripDoesNotPersistSourceOrSensitiveMarkers(t *testing.T) {
	root := t.TempDir()
	service := newTestConfigService(root, JSONConfigFileStore{}, map[string]string{
		EnvTheme:       string(ThemeDark),
		EnvDiagnostics: "true",
	})
	nextLanguage := LanguageEn
	nextDataDir := filepath.Join(root, "saved-data")

	updated, issues, err := service.Update(UpdateSettingsInput{
		Appearance: &UpdateAppearanceInput{
			Language: &nextLanguage,
		},
		Paths: &UpdatePathInput{
			DataDir: &nextDataDir,
		},
	})

	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if len(issues) != 0 {
		t.Fatalf("Update() issues = %+v, want none", issues)
	}
	reloaded, err := service.Current()
	if err != nil {
		t.Fatalf("Current() error = %v, want nil", err)
	}
	if reloaded != updated {
		t.Fatalf("Current() after save = %+v, want updated view %+v", reloaded, updated)
	}

	raw, err := os.ReadFile(filepath.Join(root, "config-root", "LoomiDBX", "config.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	serialized := string(raw)
	for _, forbidden := range []string{
		"configFile",
		"env",
		"source",
		"sensitive",
		"credential",
		"api" + "Key",
		"tok" + "en",
		"LOOMIDBX_",
		string(ThemeDark),
	} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("saved config contains forbidden marker %q: %s", forbidden, serialized)
		}
	}
	if !strings.Contains(serialized, `"language": "en"`) {
		t.Fatalf("saved config missing explicit user value: %s", serialized)
	}
	assertPathEqual(t, reloaded.Paths.DataDir, nextDataDir)
}

func assertConfigErrorDoesNotContain(t *testing.T, err ConfigError, forbidden ...string) {
	t.Helper()

	serialized := err.Error()
	for _, issue := range err.Issues {
		serialized += "\n" + issue.Message
	}
	for _, value := range forbidden {
		if strings.Contains(serialized, value) {
			t.Fatalf("ConfigError leaked forbidden value %q: %+v", value, err)
		}
	}
}

// localStorageStrategyProbe is a test substitute for the adjacent local storage strategy.
type localStorageStrategyProbe struct {
	// schemaCreated records whether the substitute attempted SQLite schema creation.
	schemaCreated bool

	// migrationsRun records whether the substitute attempted SQLite migrations.
	migrationsRun bool
}

// localStorageProbeConfig is the minimal config shape a future storage spec can derive from SettingsView.
type localStorageProbeConfig struct {
	// DataDir is copied from SettingsView.paths.dataDir.
	DataDir string

	// Mode is copied from SettingsView.development.mode.
	Mode RunMode
}

// ConfigureFromSettings consumes only SettingsView path and mode outputs for local storage placement.
func (probe *localStorageStrategyProbe) ConfigureFromSettings(view SettingsView) (localStorageProbeConfig, error) {
	if view.Paths.DataDir == "" {
		return localStorageProbeConfig{}, ConfigError{
			Code:    ConfigIssueCodeConfigPathInvalid,
			Message: "配置数据目录为空",
			Issues: []ConfigIssue{{
				Path:     "paths.dataDir",
				Code:     ConfigIssueCodeConfigPathInvalid,
				Severity: ConfigIssueSeverityError,
				Message:  "配置数据目录为空",
			}},
		}
	}
	if view.Development.Mode == "" {
		return localStorageProbeConfig{}, ConfigError{
			Code:    ConfigIssueCodeValidationFailed,
			Message: "运行模式为空",
			Issues: []ConfigIssue{{
				Path:     "development.mode",
				Code:     ConfigIssueCodeValidationFailed,
				Severity: ConfigIssueSeverityError,
				Message:  "运行模式为空",
			}},
		}
	}
	return localStorageProbeConfig{
		DataDir: view.Paths.DataDir,
		Mode:    view.Development.Mode,
	}, nil
}

func readConfigReadme(t *testing.T) string {
	t.Helper()

	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("ReadFile(internal/config/README.md) error = %v", err)
	}
	return string(raw)
}

func joinIssueMessages(issues []ConfigIssue) string {
	var builder strings.Builder
	for _, issue := range issues {
		builder.WriteString(issue.Message)
		builder.WriteByte('\n')
	}
	return builder.String()
}
