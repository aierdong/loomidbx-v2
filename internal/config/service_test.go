package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigServiceLoadReturnsComposedValidatedConfig(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config-root", "LoomiDBX", "config.json")
	writeRawFile(t, configPath, `{
		"version": 1,
		"appearance": {"language": "en", "theme": "light"},
		"development": {"mode": "desktop", "diagnosticsEnabled": false}
	}`)
	service := newTestConfigService(root, JSONConfigFileStore{}, map[string]string{
		EnvTheme:       string(ThemeDark),
		EnvDiagnostics: "true",
	})

	result, err := service.Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if result.Config.Appearance.Language != LanguageEn {
		t.Fatalf("Language = %q, want file value %q", result.Config.Appearance.Language, LanguageEn)
	}
	if result.Config.Appearance.Theme != ThemeDark {
		t.Fatalf("Theme = %q, want env override %q", result.Config.Appearance.Theme, ThemeDark)
	}
	if !result.Config.Development.DiagnosticsEnabled {
		t.Fatal("DiagnosticsEnabled = false, want env override true")
	}
	assertPathEqual(t, result.Config.Paths.ConfigFile, configPath)
	if result.Source.FileState != FileStatePresent || !result.Source.FileLoaded {
		t.Fatalf("Source = %+v, want present loaded file", result.Source)
	}
}

func TestConfigServiceCurrentReturnsSettingsView(t *testing.T) {
	root := t.TempDir()
	service := newTestConfigService(root, JSONConfigFileStore{}, nil)

	view, err := service.Current()

	if err != nil {
		t.Fatalf("Current() error = %v, want nil", err)
	}
	if view.Appearance.Language != LanguageZh || view.Appearance.Theme != ThemeSystem {
		t.Fatalf("Appearance = %+v, want default settings view", view.Appearance)
	}
	assertPathEqual(t, view.Paths.ConfigFile, filepath.Join(root, "config-root", "LoomiDBX", "config.json"))
	assertPathEqual(t, view.Paths.DataDir, filepath.Join(root, "data-root", "LoomiDBX", "data"))
}

func TestConfigServiceUpdateOnlyChangesSpecifiedFieldsAndPreservesExistingValues(t *testing.T) {
	root := t.TempDir()
	existingDataDir := filepath.Join(root, "existing-data")
	configPath := filepath.Join(root, "config-root", "LoomiDBX", "config.json")
	writeRawFile(t, configPath, `{
		"version": 1,
		"appearance": {"language": "en", "theme": "light"},
		"paths": {"dataDir": "`+jsonPath(existingDataDir)+`"},
		"development": {"mode": "development", "useIsolatedDataDir": true, "diagnosticsEnabled": false},
		"privacy": {"localOnly": true, "telemetryEnabled": false}
	}`)
	service := newTestConfigService(root, JSONConfigFileStore{}, nil)
	nextTheme := ThemeDark

	view, issues, err := service.Update(UpdateSettingsInput{
		Appearance: &UpdateAppearanceInput{
			Theme: &nextTheme,
		},
	})

	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if len(issues) != 0 {
		t.Fatalf("Update() issues = %+v, want none", issues)
	}
	if view.Appearance.Theme != ThemeDark || view.Appearance.Language != LanguageEn {
		t.Fatalf("Appearance = %+v, want theme changed and language preserved", view.Appearance)
	}
	assertPathEqual(t, view.Paths.DataDir, existingDataDir)
	if view.Development.Mode != ModeDevelopment || !view.Development.UseIsolatedDataDir || view.Development.DiagnosticsEnabled {
		t.Fatalf("Development = %+v, want existing values preserved", view.Development)
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	serialized := string(raw)
	for _, forbidden := range nonPersistentOrSensitiveMarkers() {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("saved config contains non-persistent or sensitive marker %q: %s", forbidden, serialized)
		}
	}
	if !strings.Contains(serialized, `"theme": "dark"`) {
		t.Fatalf("saved config does not contain updated theme: %s", serialized)
	}
	if !strings.Contains(serialized, `"language": "en"`) {
		t.Fatalf("saved config did not preserve existing language: %s", serialized)
	}
}

func nonPersistentOrSensitiveMarkers() []string {
	return []string{
		"configFile",
		"sources",
		"LOOMIDBX_",
		"Plaintext",
		"api" + "Key",
		"tok" + "en",
	}
}

func TestConfigServiceUpdateReloadsSavedUserSettingsConsistently(t *testing.T) {
	root := t.TempDir()
	service := newTestConfigService(root, JSONConfigFileStore{}, nil)
	nextLanguage := LanguageEn
	nextTheme := ThemeDark
	nextDataDir := filepath.Join(root, "custom-data")
	diagnostics := true

	updated, issues, err := service.Update(UpdateSettingsInput{
		Appearance: &UpdateAppearanceInput{
			Language: &nextLanguage,
			Theme:    &nextTheme,
		},
		Paths: &UpdatePathInput{
			DataDir: &nextDataDir,
		},
		Development: &UpdateDevelopmentInput{
			DiagnosticsEnabled: &diagnostics,
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
		t.Fatalf("Current() after Update() error = %v, want nil", err)
	}
	if reloaded != updated {
		t.Fatalf("Current() after save = %+v, want updated view %+v", reloaded, updated)
	}
	if reloaded.Appearance.Language != LanguageEn || reloaded.Appearance.Theme != ThemeDark {
		t.Fatalf("Appearance = %+v, want saved user values", reloaded.Appearance)
	}
	assertPathEqual(t, reloaded.Paths.DataDir, nextDataDir)
	if !reloaded.Development.DiagnosticsEnabled {
		t.Fatal("DiagnosticsEnabled = false, want saved true")
	}
}

func TestConfigServiceUpdatePreservesDefaultValuesForNewSections(t *testing.T) {
	root := t.TempDir()
	service := newTestConfigService(root, JSONConfigFileStore{}, nil)
	telemetry := true

	view, issues, err := service.Update(UpdateSettingsInput{
		Privacy: &UpdatePrivacyInput{
			TelemetryEnabled: &telemetry,
		},
	})

	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if len(issues) != 0 {
		t.Fatalf("Update() issues = %+v, want none", issues)
	}
	if !view.Privacy.LocalOnly {
		t.Fatal("LocalOnly = false, want default true preserved")
	}
	if !view.Privacy.TelemetryEnabled {
		t.Fatal("TelemetryEnabled = false, want updated true")
	}
}

func TestConfigServiceUpdateDoesNotPersistEnvironmentOverrides(t *testing.T) {
	root := t.TempDir()
	service := newTestConfigService(root, JSONConfigFileStore{}, map[string]string{
		EnvTheme: string(ThemeDark),
	})
	nextLanguage := LanguageEn

	_, issues, err := service.Update(UpdateSettingsInput{
		Appearance: &UpdateAppearanceInput{
			Language: &nextLanguage,
		},
	})

	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if len(issues) != 0 {
		t.Fatalf("Update() issues = %+v, want none", issues)
	}
	raw, readErr := os.ReadFile(filepath.Join(root, "config-root", "LoomiDBX", "config.json"))
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	serialized := string(raw)
	if strings.Contains(serialized, string(ThemeDark)) {
		t.Fatalf("env theme was persisted: %s", serialized)
	}
	if !strings.Contains(serialized, `"language": "en"`) {
		t.Fatalf("explicit update missing from persisted config: %s", serialized)
	}
}

func TestConfigServiceUpdateAllowsPartialExistingUserConfig(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config-root", "LoomiDBX", "config.json")
	writeRawFile(t, configPath, `{
		"version": 1,
		"appearance": {"language": "zh"}
	}`)
	service := newTestConfigService(root, JSONConfigFileStore{}, nil)
	nextLanguage := LanguageEn

	view, issues, err := service.Update(UpdateSettingsInput{
		Appearance: &UpdateAppearanceInput{
			Language: &nextLanguage,
		},
	})

	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if len(issues) != 0 {
		t.Fatalf("Update() issues = %+v, want none", issues)
	}
	if view.Appearance.Language != LanguageEn || view.Appearance.Theme != ThemeSystem {
		t.Fatalf("Appearance = %+v, want updated language and default theme", view.Appearance)
	}
}

func TestConfigServiceInvalidUpdateReturnsIssueAndDoesNotWriteFile(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config-root", "LoomiDBX", "config.json")
	writeRawFile(t, configPath, `{
		"version": 1,
		"appearance": {"language": "zh", "theme": "system"}
	}`)
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(before) error = %v", err)
	}
	service := newTestConfigService(root, JSONConfigFileStore{}, nil)
	invalidTheme := Theme("blue")

	view, issues, err := service.Update(UpdateSettingsInput{
		Appearance: &UpdateAppearanceInput{
			Theme: &invalidTheme,
		},
	})

	if err != nil {
		t.Fatalf("Update() error = %v, want nil validation result", err)
	}
	if view != (SettingsView{}) {
		t.Fatalf("Update() view = %+v, want zero view for invalid update", view)
	}
	assertIssue(t, issues, "appearance.theme", ConfigIssueCodeValidationFailed)
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(after) error = %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("config file changed after invalid update:\nbefore=%s\nafter=%s", before, after)
	}
}

func TestConfigServiceSaveFailureReturnsReason(t *testing.T) {
	root := t.TempDir()
	store := failingSaveStore{
		delegate: JSONConfigFileStore{},
		err:      errors.New("simulated write failure at C:/Users/example/secret-path"),
	}
	service := newTestConfigService(root, store, nil)
	nextTheme := ThemeDark

	_, issues, err := service.Update(UpdateSettingsInput{
		Appearance: &UpdateAppearanceInput{
			Theme: &nextTheme,
		},
	})

	if err == nil {
		t.Fatal("Update() error = nil, want write failure")
	}
	configErr, ok := err.(ConfigError)
	if !ok {
		t.Fatalf("Update() error = %T %[1]v, want ConfigError", err)
	}
	if configErr.Code != ConfigIssueCodeConfigWriteFailed {
		t.Fatalf("Code = %q, want %q", configErr.Code, ConfigIssueCodeConfigWriteFailed)
	}
	assertIssue(t, configErr.Issues, "paths.configFile", ConfigIssueCodeConfigWriteFailed)
	if len(issues) != 0 {
		t.Fatalf("Update() issues = %+v, want none when write failure is returned as error", issues)
	}
	if strings.Contains(configErr.Message, "C:/") || strings.Contains(configErr.Message, "secret-path") {
		t.Fatalf("Message = %q, want sanitized write failure reason", configErr.Message)
	}
}

func newTestConfigService(root string, store ConfigFileStore, env map[string]string) ConfigService {
	return NewConfigService(ConfigServiceOptions{
		Loader: NewConfigLoader(ConfigLoaderOptions{
			Store:     store,
			Resolver:  DefaultPathResolver{},
			EnvReader: mapEnvReader(env),
			PathInput: PathResolveInput{
				AppName:           "LoomiDBX",
				Mode:              ModeDesktop,
				DesktopConfigRoot: filepath.Join(root, "config-root"),
				DesktopDataRoot:   filepath.Join(root, "data-root"),
			},
		}),
	})
}

// failingSaveStore delegates reads and returns a configured write error for save failure tests.
type failingSaveStore struct {
	// delegate handles Read calls using the same behavior as the real store.
	delegate ConfigFileStore

	// err is returned by Save to simulate a persistence failure.
	err error
}

func (store failingSaveStore) Read(path string) (UserConfig, FileState, error) {
	return store.delegate.Read(path)
}

func (store failingSaveStore) Save(path string, config UserConfig) error {
	return store.err
}
