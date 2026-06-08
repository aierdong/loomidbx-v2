package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigLoaderMissingFileReturnsCompleteResolvedDefaults(t *testing.T) {
	root := t.TempDir()
	loader := newTestConfigLoader(t, root, nil)

	result, err := loader.Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if result.Source.FileState != FileStateMissing {
		t.Fatalf("FileState = %q, want %q", result.Source.FileState, FileStateMissing)
	}
	if !result.Source.DefaultsApplied {
		t.Fatal("DefaultsApplied = false, want true")
	}
	if result.Source.FileLoaded {
		t.Fatal("FileLoaded = true, want false for missing file")
	}
	if result.Config.Version != CurrentConfigVersion {
		t.Fatalf("Version = %d, want %d", result.Config.Version, CurrentConfigVersion)
	}
	if result.Config.Appearance.Language != LanguageZh || result.Config.Appearance.Theme != ThemeSystem {
		t.Fatalf("Appearance = %+v, want default zh/system", result.Config.Appearance)
	}
	assertPathEqual(t, result.Config.Paths.ConfigFile, filepath.Join(root, "config-root", "LoomiDBX", "config.json"))
	assertPathEqual(t, result.Config.Paths.DataDir, filepath.Join(root, "data-root", "LoomiDBX", "data"))
	if strings.Contains(result.Config.Paths.ConfigFile, "{") || strings.Contains(result.Config.Paths.DataDir, "{") {
		t.Fatalf("resolved paths still contain placeholders: %+v", result.Config.Paths)
	}
}

func TestConfigLoaderFileConfigOverridesDefaults(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config-root", "LoomiDBX", "config.json")
	fileDataDir := filepath.Join(root, "file-data")
	writeRawFile(t, configPath, `{
		"version": 1,
		"appearance": {"language": "en", "theme": "dark"},
		"paths": {"dataDir": "`+jsonPath(fileDataDir)+`"},
		"development": {"mode": "development", "useIsolatedDataDir": true, "diagnosticsEnabled": true},
		"privacy": {"localOnly": false, "telemetryEnabled": true}
	}`)
	loader := newTestConfigLoader(t, root, nil)

	result, err := loader.Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if result.Source.FileState != FileStatePresent || !result.Source.FileLoaded {
		t.Fatalf("Source = %+v, want present loaded file", result.Source)
	}
	if result.Config.Appearance.Language != LanguageEn || result.Config.Appearance.Theme != ThemeDark {
		t.Fatalf("Appearance = %+v, want file en/dark", result.Config.Appearance)
	}
	assertPathEqual(t, result.Config.Paths.DataDir, fileDataDir)
	if result.Config.Development.Mode != ModeDevelopment || !result.Config.Development.UseIsolatedDataDir || !result.Config.Development.DiagnosticsEnabled {
		t.Fatalf("Development = %+v, want file development flags", result.Config.Development)
	}
	if result.Config.Privacy.LocalOnly || !result.Config.Privacy.TelemetryEnabled {
		t.Fatalf("Privacy = %+v, want file privacy values", result.Config.Privacy)
	}
}

func TestConfigLoaderFileModeReResolvesFinalPaths(t *testing.T) {
	root := t.TempDir()
	testRoot := filepath.Join(root, "test-root")
	fileDataDir := filepath.Join(root, "file-data")
	configPath := filepath.Join(root, "config-root", "LoomiDBX", "config.json")
	writeRawFile(t, configPath, `{
		"version": 1,
		"paths": {"dataDir": "`+jsonPath(fileDataDir)+`"},
		"development": {"mode": "test"}
	}`)
	loader := NewConfigLoader(ConfigLoaderOptions{
		Store:     JSONConfigFileStore{},
		Resolver:  DefaultPathResolver{},
		EnvReader: mapEnvReader(nil),
		PathInput: PathResolveInput{
			AppName:           "LoomiDBX",
			Mode:              ModeDesktop,
			TestRoot:          testRoot,
			DesktopConfigRoot: filepath.Join(root, "config-root"),
			DesktopDataRoot:   filepath.Join(root, "data-root"),
		},
	})

	result, err := loader.Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if result.Config.Development.Mode != ModeTest {
		t.Fatalf("Mode = %q, want %q", result.Config.Development.Mode, ModeTest)
	}
	if !result.Config.Development.UseIsolatedDataDir {
		t.Fatal("UseIsolatedDataDir = false, want true for test mode")
	}
	assertPathEqual(t, result.Config.Paths.ConfigFile, filepath.Join(testRoot, "config", "config.json"))
	assertPathEqual(t, result.Config.Paths.DataDir, filepath.Join(testRoot, "data"))
	assertPathEqual(t, result.Source.ConfigFile, configPath)
	assertPathEqual(t, result.Source.ResolvedPaths.ConfigFile, filepath.Join(testRoot, "config", "config.json"))
	assertPathEqual(t, result.Source.ResolvedPaths.DataDir, filepath.Join(testRoot, "data"))
}

func TestConfigLoaderEnvironmentOverridesFileValues(t *testing.T) {
	root := t.TempDir()
	fileDataDir := filepath.Join(root, "file-data")
	envDataDir := filepath.Join(root, "env-data")
	writeRawFile(t, filepath.Join(root, "config-root", "LoomiDBX", "config.json"), `{
		"version": 1,
		"appearance": {"language": "zh", "theme": "light"},
		"paths": {"dataDir": "`+jsonPath(fileDataDir)+`"},
		"development": {"mode": "desktop", "diagnosticsEnabled": false}
	}`)
	loader := newTestConfigLoader(t, root, map[string]string{
		EnvLanguage:    string(LanguageEn),
		EnvTheme:       string(ThemeDark),
		EnvDataDir:     envDataDir,
		EnvDiagnostics: "true",
	})

	result, err := loader.Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if !result.Source.EnvOverrideApplied {
		t.Fatal("EnvOverrideApplied = false, want true")
	}
	assertConfigSource(t, result.Source.EnvSources, "appearance.language", EnvLanguage)
	assertConfigSource(t, result.Source.EnvSources, "appearance.theme", EnvTheme)
	assertConfigSource(t, result.Source.EnvSources, "paths.dataDir", EnvDataDir)
	if result.Config.Appearance.Language != LanguageEn || result.Config.Appearance.Theme != ThemeDark {
		t.Fatalf("Appearance = %+v, want env en/dark", result.Config.Appearance)
	}
	assertPathEqual(t, result.Config.Paths.DataDir, envDataDir)
	if !result.Config.Development.DiagnosticsEnabled {
		t.Fatal("DiagnosticsEnabled = false, want env true")
	}
}

func TestConfigLoaderEnvironmentOverridesAreNotWrittenBack(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config-root", "LoomiDBX", "config.json")
	writeRawFile(t, configPath, `{
		"version": 1,
		"appearance": {"language": "zh", "theme": "light"}
	}`)
	loader := newTestConfigLoader(t, root, map[string]string{
		EnvTheme: string(ThemeDark),
	})

	result, err := loader.Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if result.Source.EnvWriteBack {
		t.Fatal("EnvWriteBack = true, want false")
	}
	if result.Config.Appearance.Theme != ThemeDark {
		t.Fatalf("Theme = %q, want env dark", result.Config.Appearance.Theme)
	}
	raw, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	if strings.Contains(string(raw), string(ThemeDark)) {
		t.Fatalf("config file was changed by env override: %s", string(raw))
	}
}

func TestConfigLoaderMissingFileDoesNotFailStartup(t *testing.T) {
	root := t.TempDir()
	loader := newTestConfigLoader(t, root, nil)

	result, err := loader.Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil for missing config file", err)
	}
	if result.Source.FileState != FileStateMissing {
		t.Fatalf("FileState = %q, want missing", result.Source.FileState)
	}
	if result.Config == (AppConfig{}) {
		t.Fatal("Config is zero, want complete defaults")
	}
}

func TestConfigLoaderLoadResultBuildsSettingsView(t *testing.T) {
	root := t.TempDir()
	loader := newTestConfigLoader(t, root, map[string]string{
		EnvTheme: string(ThemeDark),
	})

	result, err := loader.Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	view := result.SettingsView()
	if view.Appearance.Theme != ThemeDark {
		t.Fatalf("SettingsView theme = %q, want %q", view.Appearance.Theme, ThemeDark)
	}
	assertPathEqual(t, view.Paths.ConfigFile, result.Config.Paths.ConfigFile)
	assertPathEqual(t, view.Paths.DataDir, result.Config.Paths.DataDir)
	if view.Development.Mode != result.Config.Development.Mode {
		t.Fatalf("SettingsView mode = %q, want %q", view.Development.Mode, result.Config.Development.Mode)
	}
}

func newTestConfigLoader(t *testing.T, root string, env map[string]string) ConfigLoader {
	t.Helper()
	return NewConfigLoader(ConfigLoaderOptions{
		Store:     JSONConfigFileStore{},
		Resolver:  DefaultPathResolver{},
		EnvReader: mapEnvReader(env),
		PathInput: PathResolveInput{
			AppName:           "LoomiDBX",
			Mode:              ModeDesktop,
			DesktopConfigRoot: filepath.Join(root, "config-root"),
			DesktopDataRoot:   filepath.Join(root, "data-root"),
		},
	})
}

func jsonPath(path string) string {
	return strings.ReplaceAll(path, `\`, `\\`)
}

func assertConfigSource(t *testing.T, sources []ConfigOverrideSource, path string, envName string) {
	t.Helper()
	for _, source := range sources {
		if source.Path == path && source.EnvName == envName {
			return
		}
	}
	t.Fatalf("sources = %+v, want %s from %s", sources, path, envName)
}
