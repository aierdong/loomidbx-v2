package config

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultAppConfigIsCompleteWithoutConfigFile(t *testing.T) {
	cfg := DefaultAppConfig()

	if cfg.Version != CurrentConfigVersion {
		t.Fatalf("Version = %d, want %d", cfg.Version, CurrentConfigVersion)
	}
	if cfg.Appearance.Language != LanguageZh {
		t.Fatalf("Appearance.Language = %q, want %q", cfg.Appearance.Language, LanguageZh)
	}
	if cfg.Appearance.Theme != ThemeSystem {
		t.Fatalf("Appearance.Theme = %q, want %q", cfg.Appearance.Theme, ThemeSystem)
	}
	if cfg.Paths.DataDir == "" {
		t.Fatal("Paths.DataDir must have a deterministic default placeholder")
	}
	if cfg.Paths.ConfigFile == "" {
		t.Fatal("Paths.ConfigFile must have a deterministic default placeholder")
	}
	if cfg.Development.Mode != ModeDesktop {
		t.Fatalf("Development.Mode = %q, want %q", cfg.Development.Mode, ModeDesktop)
	}
	if cfg.Development.UseIsolatedDataDir {
		t.Fatal("Development.UseIsolatedDataDir must default to false for desktop mode")
	}
	if cfg.Integrations.Account.Status != FutureStatusUnavailable {
		t.Fatalf("Account.Status = %q, want %q", cfg.Integrations.Account.Status, FutureStatusUnavailable)
	}
	if cfg.Integrations.Account.Enabled || cfg.Integrations.Account.Configured {
		t.Fatal("Account integration must default to unavailable and not configured")
	}
	if cfg.Integrations.LLM.Status != FutureStatusUnavailable {
		t.Fatalf("LLM.Status = %q, want %q", cfg.Integrations.LLM.Status, FutureStatusUnavailable)
	}
	if cfg.Integrations.LLM.Enabled || cfg.Integrations.LLM.Configured {
		t.Fatal("LLM integration must default to unavailable and not configured")
	}
	if !cfg.Privacy.LocalOnly {
		t.Fatal("Privacy.LocalOnly must default to true")
	}
	if cfg.Privacy.TelemetryEnabled {
		t.Fatal("Privacy.TelemetryEnabled must default to false")
	}
	if !cfg.Privacy.SensitiveCredentials.ExternalStorageRequired {
		t.Fatal("sensitive credentials must require an external secure storage boundary")
	}
	if cfg.Privacy.BusinessData.StoredInAppConfig {
		t.Fatal("business data must not be modeled as ordinary app config")
	}
}

func TestDefaultUserConfigSerializesOnlyOrdinaryConfig(t *testing.T) {
	userCfg := DefaultUserConfig()

	raw, err := json.Marshal(userCfg)
	if err != nil {
		t.Fatalf("Marshal(DefaultUserConfig()) error = %v", err)
	}

	serialized := string(raw)
	for _, forbidden := range []string{
		"password",
		"passwd",
		"token",
		"apiKey",
		"api_key",
		"secret",
		"llmApiKey",
		"databasePassword",
		"connectionString",
		"userSQL",
		"schemaCache",
		"projectConfig",
		"generatedData",
	} {
		if strings.Contains(strings.ToLower(serialized), strings.ToLower(forbidden)) {
			t.Fatalf("serialized user config contains forbidden sensitive/business field %q: %s", forbidden, serialized)
		}
	}
}

func TestPersistentUserConfigDoesNotExposePlaintextSecretFields(t *testing.T) {
	assertNoPlaintextSecretFields(t, reflect.TypeOf(UserConfig{}))
}

func assertNoPlaintextSecretFields(t *testing.T, typ reflect.Type) {
	t.Helper()

	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := strings.ToLower(field.Name + " " + field.Tag.Get("json"))
		for _, forbidden := range []string{"password", "passwd", "token", "apikey", "api_key", "secret", "connectionstring"} {
			if strings.Contains(fieldName, forbidden) {
				t.Fatalf("UserConfig exposes plaintext-like field %s with json tag %q", field.Name, field.Tag.Get("json"))
			}
		}
		assertNoPlaintextSecretFields(t, field.Type)
	}
}
