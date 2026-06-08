package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvOverridesReadsFixedSupportedNames(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	dataDir := filepath.Join(root, "data")
	testRoot := filepath.Join(root, "test-root")
	developmentRoot := filepath.Join(root, "development-root")

	overrides, issues := ReadEnvOverrides(mapEnvReader(map[string]string{
		"LOOMIDBX_CONFIG_DIR":       configDir,
		"LOOMIDBX_DATA_DIR":         dataDir,
		"LOOMIDBX_MODE":             string(ModeTest),
		"LOOMIDBX_LANGUAGE":         string(LanguageEn),
		"LOOMIDBX_THEME":            string(ThemeDark),
		"LOOMIDBX_TEST_ROOT":        testRoot,
		"LOOMIDBX_DEVELOPMENT_ROOT": developmentRoot,
		"LOOMIDBX_DIAGNOSTICS":      "true",
	}), DefaultPathResolver{})

	assertNoEnvIssues(t, issues)
	if overrides.Appearance == nil {
		t.Fatal("Appearance overrides are nil")
	}
	if overrides.Appearance.Language == nil || *overrides.Appearance.Language != LanguageEn {
		t.Fatalf("Language override = %v, want %q", overrides.Appearance.Language, LanguageEn)
	}
	if overrides.Appearance.Theme == nil || *overrides.Appearance.Theme != ThemeDark {
		t.Fatalf("Theme override = %v, want %q", overrides.Appearance.Theme, ThemeDark)
	}
	if overrides.Development == nil {
		t.Fatal("Development overrides are nil")
	}
	if overrides.Development.Mode == nil || *overrides.Development.Mode != ModeTest {
		t.Fatalf("Mode override = %v, want %q", overrides.Development.Mode, ModeTest)
	}
	if overrides.Development.DiagnosticsEnabled == nil || !*overrides.Development.DiagnosticsEnabled {
		t.Fatalf("Diagnostics override = %v, want true", overrides.Development.DiagnosticsEnabled)
	}
	if overrides.Paths == nil {
		t.Fatal("Path overrides are nil")
	}
	assertPathEqual(t, overrides.Paths.ConfigDir, configDir)
	assertPathEqual(t, overrides.Paths.DataDir, dataDir)
	assertPathEqual(t, overrides.Paths.TestRoot, testRoot)
	assertPathEqual(t, overrides.Paths.DevelopmentRoot, developmentRoot)
	if overrides.WriteBack {
		t.Fatal("environment overrides must not be marked for write-back")
	}
	assertEnvSource(t, overrides.Sources, "appearance.language", "LOOMIDBX_LANGUAGE")
	assertEnvSource(t, overrides.Sources, "appearance.theme", "LOOMIDBX_THEME")
	assertEnvSource(t, overrides.Sources, "development.mode", "LOOMIDBX_MODE")
	assertEnvSource(t, overrides.Sources, "development.diagnosticsEnabled", "LOOMIDBX_DIAGNOSTICS")
	assertEnvSource(t, overrides.Sources, "paths.configDir", "LOOMIDBX_CONFIG_DIR")
	assertEnvSource(t, overrides.Sources, "paths.dataDir", "LOOMIDBX_DATA_DIR")
	assertEnvSource(t, overrides.Sources, "paths.testRoot", "LOOMIDBX_TEST_ROOT")
	assertEnvSource(t, overrides.Sources, "paths.developmentRoot", "LOOMIDBX_DEVELOPMENT_ROOT")
}

func TestEnvOverridesAcceptsSupportedEnumValues(t *testing.T) {
	for _, tc := range []struct {
		name     string
		envName  string
		value    string
		assertFn func(*testing.T, EnvOverrides)
	}{
		{
			name:    "desktop mode",
			envName: "LOOMIDBX_MODE",
			value:   string(ModeDesktop),
			assertFn: func(t *testing.T, overrides EnvOverrides) {
				t.Helper()
				if overrides.Development == nil || overrides.Development.Mode == nil || *overrides.Development.Mode != ModeDesktop {
					t.Fatalf("mode override = %+v, want %q", overrides.Development, ModeDesktop)
				}
			},
		},
		{
			name:    "system theme",
			envName: "LOOMIDBX_THEME",
			value:   string(ThemeSystem),
			assertFn: func(t *testing.T, overrides EnvOverrides) {
				t.Helper()
				if overrides.Appearance == nil || overrides.Appearance.Theme == nil || *overrides.Appearance.Theme != ThemeSystem {
					t.Fatalf("theme override = %+v, want %q", overrides.Appearance, ThemeSystem)
				}
			},
		},
		{
			name:    "zh language",
			envName: "LOOMIDBX_LANGUAGE",
			value:   string(LanguageZh),
			assertFn: func(t *testing.T, overrides EnvOverrides) {
				t.Helper()
				if overrides.Appearance == nil || overrides.Appearance.Language == nil || *overrides.Appearance.Language != LanguageZh {
					t.Fatalf("language override = %+v, want %q", overrides.Appearance, LanguageZh)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			overrides, issues := ReadEnvOverrides(mapEnvReader(map[string]string{tc.envName: tc.value}), DefaultPathResolver{})

			assertNoEnvIssues(t, issues)
			tc.assertFn(t, overrides)
		})
	}
}

func TestEnvOverridesReportsInvalidEnumValues(t *testing.T) {
	_, issues := ReadEnvOverrides(mapEnvReader(map[string]string{
		"LOOMIDBX_MODE":     "server",
		"LOOMIDBX_LANGUAGE": "fr",
		"LOOMIDBX_THEME":    "blue",
	}), DefaultPathResolver{})

	assertIssue(t, issues, "development.mode", ConfigIssueCodeValidationFailed)
	assertIssue(t, issues, "appearance.language", ConfigIssueCodeValidationFailed)
	assertIssue(t, issues, "appearance.theme", ConfigIssueCodeValidationFailed)
}

func TestEnvOverridesPassesAbsolutePathsThroughPathResolver(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "env-config")
	dataDir := filepath.Join(root, "env-data")

	overrides, issues := ReadEnvOverrides(mapEnvReader(map[string]string{
		"LOOMIDBX_CONFIG_DIR": configDir,
		"LOOMIDBX_DATA_DIR":   dataDir,
	}), DefaultPathResolver{})

	assertNoEnvIssues(t, issues)
	if overrides.ResolvedPaths == nil {
		t.Fatal("ResolvedPaths is nil")
	}
	assertPathEqual(t, overrides.ResolvedPaths.ConfigDir, configDir)
	assertPathEqual(t, overrides.ResolvedPaths.ConfigFile, filepath.Join(configDir, "config.json"))
	assertPathEqual(t, overrides.ResolvedPaths.DataDir, dataDir)
}

func TestEnvOverridesReportsNonAbsolutePathIssues(t *testing.T) {
	_, issues := ReadEnvOverrides(mapEnvReader(map[string]string{
		"LOOMIDBX_CONFIG_DIR":       "relative-config",
		"LOOMIDBX_DATA_DIR":         "relative-data",
		"LOOMIDBX_TEST_ROOT":        "relative-test-root",
		"LOOMIDBX_DEVELOPMENT_ROOT": "relative-development-root",
	}), DefaultPathResolver{})

	assertIssue(t, issues, "paths.configDir", ConfigIssueCodeConfigPathInvalid)
	assertIssue(t, issues, "paths.dataDir", ConfigIssueCodeConfigPathInvalid)
	assertIssue(t, issues, "paths.testRoot", ConfigIssueCodeConfigPathInvalid)
	assertIssue(t, issues, "paths.developmentRoot", ConfigIssueCodeConfigPathInvalid)
}

func TestEnvOverridesRejectsSensitiveFixedNamesWithoutLeakingValue(t *testing.T) {
	secret := "super-secret-value"

	_, issues := ReadEnvOverrides(mapEnvReader(map[string]string{
		"LOOMIDBX_API_KEY":     secret,
		"LOOMIDBX_TOKEN":       secret,
		"LOOMIDBX_PASSWORD":    secret,
		"LOOMIDBX_LLM_API_KEY": secret,
	}), DefaultPathResolver{})

	assertIssue(t, issues, "env.LOOMIDBX_API_KEY", ConfigIssueCodeSensitiveValueNotAllowed)
	assertIssue(t, issues, "env.LOOMIDBX_TOKEN", ConfigIssueCodeSensitiveValueNotAllowed)
	assertIssue(t, issues, "env.LOOMIDBX_PASSWORD", ConfigIssueCodeSensitiveValueNotAllowed)
	assertIssue(t, issues, "env.LOOMIDBX_LLM_API_KEY", ConfigIssueCodeSensitiveValueNotAllowed)
	assertMessagesDoNotContain(t, issues, secret)
}

func TestEnvOverridesCanApplyOverFileConfigWithoutWriteBack(t *testing.T) {
	language := LanguageEn
	theme := ThemeDark
	mode := ModeDevelopment
	diagnosticsEnabled := true
	configDir := filepath.Join(t.TempDir(), "config")
	dataDir := filepath.Join(t.TempDir(), "data")
	fileConfig := UserConfig{
		Appearance: &AppearanceConfig{
			Language: LanguageZh,
			Theme:    ThemeLight,
		},
		Paths: &UserPathConfig{
			DataDir: filepath.Join(t.TempDir(), "file-data"),
		},
		Development: &UserDevelopmentConfig{
			Mode:               ModeDesktop,
			DiagnosticsEnabled: boolPtr(false),
		},
	}
	overrides := EnvOverrides{
		Appearance: &EnvAppearanceOverrides{
			Language: &language,
			Theme:    &theme,
		},
		Paths: &EnvPathOverrides{
			ConfigDir: configDir,
			DataDir:   dataDir,
		},
		Development: &EnvDevelopmentOverrides{
			Mode:               &mode,
			DiagnosticsEnabled: &diagnosticsEnabled,
		},
		WriteBack: false,
	}

	applied := overrides.ApplyToUserConfig(fileConfig)

	if applied.Appearance == nil || applied.Appearance.Language != LanguageEn || applied.Appearance.Theme != ThemeDark {
		t.Fatalf("Appearance after apply = %+v, want env values", applied.Appearance)
	}
	if applied.Paths == nil || applied.Paths.DataDir != dataDir {
		t.Fatalf("Paths after apply = %+v, want dataDir %q", applied.Paths, dataDir)
	}
	if applied.Development == nil || applied.Development.Mode != ModeDevelopment || applied.Development.DiagnosticsEnabled == nil || !*applied.Development.DiagnosticsEnabled {
		t.Fatalf("Development after apply = %+v, want env values", applied.Development)
	}
	if overrides.WriteBack {
		t.Fatal("environment overrides must remain non-persistent after apply")
	}
	if fileConfig.Appearance == nil || fileConfig.Appearance.Language != LanguageZh || fileConfig.Appearance.Theme != ThemeLight {
		t.Fatalf("source Appearance was mutated = %+v", fileConfig.Appearance)
	}
	if fileConfig.Paths == nil || fileConfig.Paths.DataDir == dataDir {
		t.Fatalf("source Paths was mutated = %+v", fileConfig.Paths)
	}
	if fileConfig.Development == nil || fileConfig.Development.Mode != ModeDesktop || fileConfig.Development.DiagnosticsEnabled == nil || *fileConfig.Development.DiagnosticsEnabled {
		t.Fatalf("source Development was mutated = %+v", fileConfig.Development)
	}
}

func TestEnvOverridesApplyUsesResolvedTestDataDir(t *testing.T) {
	root := t.TempDir()
	testRoot := filepath.Join(root, "test-root")
	rawDataOverride := filepath.Join(root, "raw-data")
	fileDataDir := filepath.Join(root, "file-data")

	overrides, issues := ReadEnvOverrides(mapEnvReader(map[string]string{
		"LOOMIDBX_MODE":      string(ModeTest),
		"LOOMIDBX_TEST_ROOT": testRoot,
		"LOOMIDBX_DATA_DIR":  rawDataOverride,
	}), DefaultPathResolver{})

	assertNoEnvIssues(t, issues)
	if overrides.ResolvedPaths == nil {
		t.Fatal("ResolvedPaths is nil")
	}
	applied := overrides.ApplyToUserConfig(UserConfig{
		Paths: &UserPathConfig{DataDir: fileDataDir},
	})

	assertPathEqual(t, overrides.ResolvedPaths.DataDir, filepath.Join(testRoot, "data"))
	if applied.Paths == nil {
		t.Fatal("applied Paths is nil")
	}
	assertPathEqual(t, applied.Paths.DataDir, filepath.Join(testRoot, "data"))
	if filepath.Clean(applied.Paths.DataDir) == filepath.Clean(rawDataOverride) {
		t.Fatal("test mode data override bypassed resolver isolation")
	}
}

func mapEnvReader(values map[string]string) EnvReader {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}

func assertNoEnvIssues(t *testing.T, issues []ConfigIssue) {
	t.Helper()
	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func assertEnvSource(t *testing.T, sources []ConfigOverrideSource, path string, envName string) {
	t.Helper()
	for _, source := range sources {
		if source.Path == path && source.EnvName == envName {
			return
		}
	}
	t.Fatalf("sources = %+v, want %s from %s", sources, path, envName)
}

func TestEnvOverrideMessagesAvoidRawInvalidValues(t *testing.T) {
	rawValue := "invalid-theme-value"

	_, issues := ReadEnvOverrides(mapEnvReader(map[string]string{
		"LOOMIDBX_THEME": rawValue,
	}), DefaultPathResolver{})

	assertIssue(t, issues, "appearance.theme", ConfigIssueCodeValidationFailed)
	for _, issue := range issues {
		if strings.Contains(issue.Message, rawValue) {
			t.Fatalf("issue message %q contains raw invalid value %q", issue.Message, rawValue)
		}
	}
}
