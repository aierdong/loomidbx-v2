package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigFileStoreReadMissingReturnsMissingWithoutError(t *testing.T) {
	store := JSONConfigFileStore{}
	path := filepath.Join(t.TempDir(), "config", "config.json")

	cfg, state, err := store.Read(path)

	if err != nil {
		t.Fatalf("Read() error = %v, want nil", err)
	}
	if state != FileStateMissing {
		t.Fatalf("state = %q, want %q", state, FileStateMissing)
	}
	if cfg != (UserConfig{}) {
		t.Fatalf("cfg = %+v, want zero UserConfig", cfg)
	}
}

func TestConfigFileStoreReadValidUserConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	writeRawFile(t, path, `{
		"version": 1,
		"appearance": {"language": "en", "theme": "dark"},
		"paths": {"dataDir": "E:/loomidbx/data"},
		"development": {"mode": "development", "useIsolatedDataDir": true, "diagnosticsEnabled": true},
		"integrations": {
			"account": {"enabled": false, "configured": false, "status": "unavailable"},
			"llm": {"enabled": false, "configured": true, "status": "not_configured"}
		},
		"privacy": {"localOnly": true, "telemetryEnabled": false}
	}`)

	cfg, state, err := (JSONConfigFileStore{}).Read(path)

	if err != nil {
		t.Fatalf("Read() error = %v, want nil", err)
	}
	if state != FileStatePresent {
		t.Fatalf("state = %q, want %q", state, FileStatePresent)
	}
	if cfg.Version != CurrentConfigVersion {
		t.Fatalf("Version = %d, want %d", cfg.Version, CurrentConfigVersion)
	}
	if cfg.Appearance == nil || cfg.Appearance.Language != LanguageEn || cfg.Appearance.Theme != ThemeDark {
		t.Fatalf("Appearance = %+v, want en/dark", cfg.Appearance)
	}
	if cfg.Paths == nil || cfg.Paths.DataDir != "E:/loomidbx/data" {
		t.Fatalf("Paths = %+v, want dataDir", cfg.Paths)
	}
	if cfg.Development == nil || cfg.Development.Mode != ModeDevelopment || !cfg.Development.UseIsolatedDataDir || !cfg.Development.DiagnosticsEnabled {
		t.Fatalf("Development = %+v, want development flags", cfg.Development)
	}
	if cfg.Privacy == nil || !cfg.Privacy.LocalOnly || cfg.Privacy.TelemetryEnabled {
		t.Fatalf("Privacy = %+v, want local only without telemetry", cfg.Privacy)
	}
}

func TestConfigFileStoreReadInvalidJSONReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	writeRawFile(t, path, `{"appearance":`)

	_, state, err := (JSONConfigFileStore{}).Read(path)

	if err == nil {
		t.Fatal("Read() error = nil, want invalid JSON error")
	}
	if state != FileStatePresent {
		t.Fatalf("state = %q, want %q for existing invalid file", state, FileStatePresent)
	}
}

func TestConfigFileStoreSaveCreatesParentAndCanReadAgain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config", "config.json")
	cfg := sampleUserConfig()

	if err := (JSONConfigFileStore{}).Save(path, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	readBack, state, err := (JSONConfigFileStore{}).Read(path)
	if err != nil {
		t.Fatalf("Read() after Save() error = %v", err)
	}
	if state != FileStatePresent {
		t.Fatalf("state = %q, want %q", state, FileStatePresent)
	}
	if readBack.Appearance == nil || readBack.Appearance.Language != LanguageEn || readBack.Appearance.Theme != ThemeDark {
		t.Fatalf("readBack.Appearance = %+v, want saved values", readBack.Appearance)
	}
	if readBack.Paths == nil || readBack.Paths.DataDir != cfg.Paths.DataDir {
		t.Fatalf("readBack.Paths = %+v, want %+v", readBack.Paths, cfg.Paths)
	}
}

func TestConfigFileStoreSaveOmitsResolvedConfigFileAndEnvOnlySources(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := (JSONConfigFileStore{}).Save(path, sampleUserConfig()); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	serialized := string(raw)
	for _, forbidden := range []string{
		"configFile",
		"configDir",
		"testRoot",
		"developmentRoot",
		"sources",
		"writeBack",
		"LOOMIDBX_",
	} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("saved config contains non-persistent field %q: %s", forbidden, serialized)
		}
	}
	if !json.Valid(raw) {
		t.Fatalf("saved config is not valid JSON: %s", serialized)
	}
}

func TestConfigFileStoreSaveRejectsInvalidUserConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	invalid := sampleUserConfig()
	invalid.Appearance.Theme = Theme("blue")

	err := (JSONConfigFileStore{}).Save(path, invalid)

	configErr, ok := err.(ConfigError)
	if !ok {
		t.Fatalf("Save() error = %T %[1]v, want ConfigError", err)
	}
	assertIssue(t, configErr.Issues, "appearance.theme", ConfigIssueCodeValidationFailed)
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("config file exists after rejected save: %v", statErr)
	}
}

func TestConfigFileStoreSaveFailureDoesNotLeaveInvalidTargetFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "config.json")
	writeRawFile(t, path, `{"version":1,"appearance":{"language":"zh","theme":"system"}}`)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(before) error = %v", err)
	}

	blockingDir := filepath.Join(root, "blocking-dir")
	if err := os.Mkdir(blockingDir, 0o700); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	err = (JSONConfigFileStore{}).Save(blockingDir, sampleUserConfig())

	if err == nil {
		t.Fatal("Save() error = nil, want target replace failure")
	}
	after, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile(after) error = %v", readErr)
	}
	if string(after) != string(before) {
		t.Fatalf("unrelated valid config changed after failed save: %s", string(after))
	}
	if !json.Valid(after) {
		t.Fatalf("existing config is no longer valid JSON: %s", string(after))
	}
}

func TestConfigFileStoreReplaceFailurePreservesExistingValidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	writeRawFile(t, path, `{"version":1,"appearance":{"language":"zh","theme":"system"}}`)
	store := JSONConfigFileStore{
		replace: func(oldPath string, newPath string) error {
			if oldPath == "" || newPath != path {
				t.Fatalf("replace called with oldPath=%q newPath=%q, want target %q", oldPath, newPath, path)
			}
			return errors.New("simulated replace failure")
		},
	}

	err := store.Save(path, sampleUserConfig())

	if err == nil {
		t.Fatal("Save() error = nil, want simulated replace failure")
	}
	readBack, state, readErr := (JSONConfigFileStore{}).Read(path)
	if readErr != nil {
		t.Fatalf("Read() after failed Save() error = %v", readErr)
	}
	if state != FileStatePresent {
		t.Fatalf("state = %q, want %q", state, FileStatePresent)
	}
	if readBack.Appearance == nil || readBack.Appearance.Language != LanguageZh || readBack.Appearance.Theme != ThemeSystem {
		t.Fatalf("existing config after failed Save() = %+v, want original zh/system", readBack.Appearance)
	}
}

func sampleUserConfig() UserConfig {
	return UserConfig{
		Version: CurrentConfigVersion,
		Appearance: &AppearanceConfig{
			Language: LanguageEn,
			Theme:    ThemeDark,
		},
		Paths: &UserPathConfig{
			DataDir: "E:/loomidbx/data",
		},
		Development: &UserDevelopmentConfig{
			Mode:               ModeDevelopment,
			UseIsolatedDataDir: true,
			DiagnosticsEnabled: true,
		},
		Integrations: &FutureIntegrationsConfig{
			Account: FutureIntegrationConfig{
				Enabled:    false,
				Configured: false,
				Status:     FutureStatusUnavailable,
			},
			LLM: FutureIntegrationConfig{
				Enabled:    false,
				Configured: false,
				Status:     FutureStatusUnavailable,
			},
		},
		Privacy: &UserPrivacyConfig{
			LocalOnly:        true,
			TelemetryEnabled: false,
		},
	}
}

func writeRawFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
