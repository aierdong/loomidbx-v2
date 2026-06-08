package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
