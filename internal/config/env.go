package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// EnvConfigDir overrides the directory that contains the ordinary config file.
	EnvConfigDir = "LOOMIDBX_CONFIG_DIR"

	// EnvDataDir overrides the local application data directory.
	EnvDataDir = "LOOMIDBX_DATA_DIR"

	// EnvMode overrides the runtime mode used by config loading and path resolution.
	EnvMode = "LOOMIDBX_MODE"

	// EnvLanguage overrides the user interface language.
	EnvLanguage = "LOOMIDBX_LANGUAGE"

	// EnvTheme overrides the user interface theme.
	EnvTheme = "LOOMIDBX_THEME"

	// EnvTestRoot provides an isolated root for test config and data paths.
	EnvTestRoot = "LOOMIDBX_TEST_ROOT"

	// EnvDevelopmentRoot provides an isolated root for development config and data paths.
	EnvDevelopmentRoot = "LOOMIDBX_DEVELOPMENT_ROOT"

	// EnvDiagnostics enables or disables local diagnostics in development or test runs.
	EnvDiagnostics = "LOOMIDBX_DIAGNOSTICS"

	// EnvAPIKey is rejected because plaintext API keys do not belong in ordinary config.
	EnvAPIKey = "LOOMIDBX_API_KEY"

	// EnvToken is rejected because plaintext tokens do not belong in ordinary config.
	EnvToken = "LOOMIDBX_TOKEN"

	// EnvPassword is rejected because plaintext passwords do not belong in ordinary config.
	EnvPassword = "LOOMIDBX_PASSWORD"

	// EnvLLMAPIKey is rejected because plaintext LLM API keys require a secure storage boundary.
	EnvLLMAPIKey = "LOOMIDBX_LLM_API_KEY"
)

var sensitiveEnvNames = []string{
	EnvAPIKey,
	EnvToken,
	EnvPassword,
	EnvLLMAPIKey,
}

// EnvReader reads environment variables by fixed name and reports whether a value was present.
type EnvReader func(name string) (string, bool)

// ConfigOverrideSource records which environment variable supplied a field override.
type ConfigOverrideSource struct {
	// Path is the dotted configuration field path affected by the override.
	Path string

	// EnvName is the fixed environment variable name that supplied the override.
	EnvName string
}

// EnvOverrides contains non-persistent configuration values supplied by environment variables.
type EnvOverrides struct {
	// Appearance contains language and theme overrides.
	Appearance *EnvAppearanceOverrides

	// Paths contains raw path override inputs that can be used by later loading stages.
	Paths *EnvPathOverrides

	// Development contains runtime mode and diagnostics overrides.
	Development *EnvDevelopmentOverrides

	// ResolvedPaths contains resolver-validated paths when path inputs were present.
	ResolvedPaths *ResolvedPaths

	// Sources identifies the environment source for each accepted override.
	Sources []ConfigOverrideSource

	// WriteBack is always false because environment overrides must not be persisted.
	WriteBack bool
}

// EnvAppearanceOverrides contains appearance settings supplied by environment variables.
type EnvAppearanceOverrides struct {
	// Language is the environment-provided UI language when present.
	Language *Language

	// Theme is the environment-provided UI theme when present.
	Theme *Theme
}

// EnvPathOverrides contains path settings supplied by environment variables.
type EnvPathOverrides struct {
	// ConfigDir is the environment-provided directory containing config.json.
	ConfigDir string

	// DataDir is the environment-provided local application data directory.
	DataDir string

	// TestRoot is the environment-provided isolated root for test mode.
	TestRoot string

	// DevelopmentRoot is the environment-provided isolated root for development mode.
	DevelopmentRoot string
}

// EnvDevelopmentOverrides contains development settings supplied by environment variables.
type EnvDevelopmentOverrides struct {
	// Mode is the environment-provided runtime mode when present.
	Mode *RunMode

	// DiagnosticsEnabled is the environment-provided local diagnostics toggle when present.
	DiagnosticsEnabled *bool
}

// ReadEnvOverrides reads supported LoomiDBX environment variables and returns non-persistent overrides.
//
// The reader parameter allows tests to provide isolated environment values. When reader is nil, the
// current process environment is used. The resolver parameter validates path overrides and returns path
// issues; when resolver is nil, DefaultPathResolver is used. Sensitive fixed names are rejected without
// placing their raw values in issue messages.
func ReadEnvOverrides(reader EnvReader, resolver PathResolver) (EnvOverrides, []ConfigIssue) {
	if reader == nil {
		reader = os.LookupEnv
	}
	if resolver == nil {
		resolver = DefaultPathResolver{}
	}

	overrides := EnvOverrides{WriteBack: false}
	issues := sensitiveEnvIssues(reader)
	issues = append(issues, readAppearanceEnv(reader, &overrides)...)
	issues = append(issues, readDevelopmentEnv(reader, &overrides)...)
	readPathEnv(reader, &overrides)
	pathIssues := validatePathEnv(overrides.Paths)
	issues = append(issues, pathIssues...)
	if len(pathIssues) == 0 {
		issues = append(issues, resolveEnvPaths(resolver, &overrides)...)
	}

	return overrides, issues
}

// ApplyToUserConfig returns a copy of cfg with environment override values applied.
//
// This helper exists for later loader code to compose file config and environment config while preserving
// the non-persistent WriteBack=false marker. Config directory overrides are intentionally not written into
// UserConfig because the ordinary persisted path model only stores the user-editable data directory.
func (overrides EnvOverrides) ApplyToUserConfig(cfg UserConfig) UserConfig {
	cfg = cloneUserConfig(cfg)

	if overrides.Appearance != nil {
		if cfg.Appearance == nil {
			cfg.Appearance = &AppearanceConfig{}
		}
		if overrides.Appearance.Language != nil {
			cfg.Appearance.Language = *overrides.Appearance.Language
		}
		if overrides.Appearance.Theme != nil {
			cfg.Appearance.Theme = *overrides.Appearance.Theme
		}
	}

	if dataDir := overrides.effectiveDataDir(); dataDir != "" {
		if cfg.Paths == nil {
			cfg.Paths = &UserPathConfig{}
		}
		cfg.Paths.DataDir = dataDir
	}

	if overrides.Development != nil {
		if cfg.Development == nil {
			cfg.Development = &UserDevelopmentConfig{}
		}
		if overrides.Development.Mode != nil {
			cfg.Development.Mode = *overrides.Development.Mode
		}
		if overrides.Development.DiagnosticsEnabled != nil {
			cfg.Development.DiagnosticsEnabled = boolPtr(*overrides.Development.DiagnosticsEnabled)
		}
	}

	return cfg
}

func (overrides EnvOverrides) effectiveDataDir() string {
	if overrides.ResolvedPaths != nil && overrides.ResolvedPaths.DataDir != "" {
		return overrides.ResolvedPaths.DataDir
	}
	if overrides.Paths != nil {
		return overrides.Paths.DataDir
	}
	return ""
}

func cloneUserConfig(cfg UserConfig) UserConfig {
	if cfg.Appearance != nil {
		appearance := *cfg.Appearance
		cfg.Appearance = &appearance
	}
	if cfg.Paths != nil {
		paths := *cfg.Paths
		cfg.Paths = &paths
	}
	if cfg.Development != nil {
		development := *cfg.Development
		cfg.Development = &development
	}
	if cfg.Integrations != nil {
		integrations := *cfg.Integrations
		cfg.Integrations = &integrations
	}
	if cfg.Privacy != nil {
		privacy := *cfg.Privacy
		cfg.Privacy = &privacy
	}
	return cfg
}

func sensitiveEnvIssues(reader EnvReader) []ConfigIssue {
	var issues []ConfigIssue
	for _, envName := range sensitiveEnvNames {
		if value, ok := reader(envName); ok && value != "" {
			issues = append(issues, ConfigIssue{
				Path:     "env." + envName,
				Code:     ConfigIssueCodeSensitiveValueNotAllowed,
				Severity: ConfigIssueSeverityError,
				Message:  "敏感环境变量不支持写入普通配置，请使用后续安全存储边界",
			})
		}
	}
	return issues
}

func readAppearanceEnv(reader EnvReader, overrides *EnvOverrides) []ConfigIssue {
	var issues []ConfigIssue

	if value, ok := nonEmptyEnv(reader, EnvLanguage); ok {
		language, issue := parseLanguageEnv(value)
		if issue != nil {
			issues = append(issues, *issue)
		} else {
			ensureAppearance(overrides).Language = &language
			overrides.Sources = append(overrides.Sources, ConfigOverrideSource{
				Path:    "appearance.language",
				EnvName: EnvLanguage,
			})
		}
	}

	if value, ok := nonEmptyEnv(reader, EnvTheme); ok {
		theme, issue := parseThemeEnv(value)
		if issue != nil {
			issues = append(issues, *issue)
		} else {
			ensureAppearance(overrides).Theme = &theme
			overrides.Sources = append(overrides.Sources, ConfigOverrideSource{
				Path:    "appearance.theme",
				EnvName: EnvTheme,
			})
		}
	}

	return issues
}

func readDevelopmentEnv(reader EnvReader, overrides *EnvOverrides) []ConfigIssue {
	var issues []ConfigIssue

	if value, ok := nonEmptyEnv(reader, EnvMode); ok {
		mode, issue := parseModeEnv(value)
		if issue != nil {
			issues = append(issues, *issue)
		} else {
			ensureDevelopment(overrides).Mode = &mode
			overrides.Sources = append(overrides.Sources, ConfigOverrideSource{
				Path:    "development.mode",
				EnvName: EnvMode,
			})
		}
	}

	if value, ok := nonEmptyEnv(reader, EnvDiagnostics); ok {
		enabled, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			issues = append(issues, invalidEnvValueIssue("development.diagnosticsEnabled", "环境变量必须是布尔值"))
		} else {
			ensureDevelopment(overrides).DiagnosticsEnabled = &enabled
			overrides.Sources = append(overrides.Sources, ConfigOverrideSource{
				Path:    "development.diagnosticsEnabled",
				EnvName: EnvDiagnostics,
			})
		}
	}

	return issues
}

func readPathEnv(reader EnvReader, overrides *EnvOverrides) {
	if value, ok := nonEmptyEnv(reader, EnvConfigDir); ok {
		ensurePaths(overrides).ConfigDir = value
		overrides.Sources = append(overrides.Sources, ConfigOverrideSource{
			Path:    "paths.configDir",
			EnvName: EnvConfigDir,
		})
	}
	if value, ok := nonEmptyEnv(reader, EnvDataDir); ok {
		ensurePaths(overrides).DataDir = value
		overrides.Sources = append(overrides.Sources, ConfigOverrideSource{
			Path:    "paths.dataDir",
			EnvName: EnvDataDir,
		})
	}
	if value, ok := nonEmptyEnv(reader, EnvTestRoot); ok {
		ensurePaths(overrides).TestRoot = value
		overrides.Sources = append(overrides.Sources, ConfigOverrideSource{
			Path:    "paths.testRoot",
			EnvName: EnvTestRoot,
		})
	}
	if value, ok := nonEmptyEnv(reader, EnvDevelopmentRoot); ok {
		ensurePaths(overrides).DevelopmentRoot = value
		overrides.Sources = append(overrides.Sources, ConfigOverrideSource{
			Path:    "paths.developmentRoot",
			EnvName: EnvDevelopmentRoot,
		})
	}
}

func resolveEnvPaths(resolver PathResolver, overrides *EnvOverrides) []ConfigIssue {
	if overrides.Paths == nil {
		return nil
	}

	mode := ModeDesktop
	if overrides.Development != nil && overrides.Development.Mode != nil {
		mode = *overrides.Development.Mode
	}
	if !shouldResolveEnvPaths(mode, overrides.Paths) {
		return nil
	}

	resolved, issues := resolver.Resolve(PathResolveInput{
		Mode:              mode,
		ConfigDirOverride: overrides.Paths.ConfigDir,
		DataDirOverride:   overrides.Paths.DataDir,
		TestRoot:          overrides.Paths.TestRoot,
		DevelopmentRoot:   overrides.Paths.DevelopmentRoot,
	})
	overrides.ResolvedPaths = &resolved

	return issues
}

func validatePathEnv(paths *EnvPathOverrides) []ConfigIssue {
	if paths == nil {
		return nil
	}

	var issues []ConfigIssue
	for _, candidate := range []struct {
		path  string
		value string
	}{
		{path: "paths.configDir", value: paths.ConfigDir},
		{path: "paths.dataDir", value: paths.DataDir},
		{path: "paths.testRoot", value: paths.TestRoot},
		{path: "paths.developmentRoot", value: paths.DevelopmentRoot},
	} {
		if candidate.value != "" && !filepath.IsAbs(candidate.value) {
			issues = append(issues, invalidPathIssue(candidate.path, "路径必须是绝对路径"))
		}
	}
	return issues
}

func shouldResolveEnvPaths(mode RunMode, paths *EnvPathOverrides) bool {
	if paths.ConfigDir != "" || paths.DataDir != "" {
		return true
	}
	if mode == ModeTest && paths.TestRoot != "" {
		return true
	}
	if mode == ModeDevelopment && paths.DevelopmentRoot != "" {
		return true
	}
	return false
}

func nonEmptyEnv(reader EnvReader, name string) (string, bool) {
	value, ok := reader(name)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}

func parseLanguageEnv(value string) (Language, *ConfigIssue) {
	language := Language(strings.TrimSpace(value))
	switch language {
	case LanguageZh, LanguageEn:
		return language, nil
	default:
		return "", ptrIssue(invalidEnvValueIssue("appearance.language", "环境变量必须是受支持的语言枚举"))
	}
}

func parseThemeEnv(value string) (Theme, *ConfigIssue) {
	theme := Theme(strings.TrimSpace(value))
	switch theme {
	case ThemeLight, ThemeDark, ThemeSystem:
		return theme, nil
	default:
		return "", ptrIssue(invalidEnvValueIssue("appearance.theme", "环境变量必须是受支持的主题枚举"))
	}
}

func parseModeEnv(value string) (RunMode, *ConfigIssue) {
	mode := RunMode(strings.TrimSpace(value))
	switch mode {
	case ModeDesktop, ModeDevelopment, ModeTest:
		return mode, nil
	default:
		return "", ptrIssue(invalidEnvValueIssue("development.mode", "环境变量必须是受支持的运行模式枚举"))
	}
}

func ensureAppearance(overrides *EnvOverrides) *EnvAppearanceOverrides {
	if overrides.Appearance == nil {
		overrides.Appearance = &EnvAppearanceOverrides{}
	}
	return overrides.Appearance
}

func ensurePaths(overrides *EnvOverrides) *EnvPathOverrides {
	if overrides.Paths == nil {
		overrides.Paths = &EnvPathOverrides{}
	}
	return overrides.Paths
}

func ensureDevelopment(overrides *EnvOverrides) *EnvDevelopmentOverrides {
	if overrides.Development == nil {
		overrides.Development = &EnvDevelopmentOverrides{}
	}
	return overrides.Development
}

func ptrIssue(issue ConfigIssue) *ConfigIssue {
	return &issue
}

func invalidEnvValueIssue(path string, message string) ConfigIssue {
	return ConfigIssue{
		Path:     path,
		Code:     ConfigIssueCodeValidationFailed,
		Severity: ConfigIssueSeverityError,
		Message:  message,
	}
}
