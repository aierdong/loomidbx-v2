package config

// ConfigLoader composes application defaults, ordinary file config, and non-persistent environment overrides.
type ConfigLoader struct {
	// Store reads the ordinary user config file.
	Store ConfigFileStore

	// Resolver resolves concrete config and data paths before file loading completes.
	Resolver PathResolver

	// EnvReader reads supported environment variables for development and test overrides.
	EnvReader EnvReader

	// PathInput carries deterministic path roots and runtime mode defaults for this loader instance.
	PathInput PathResolveInput
}

// ConfigLoaderOptions contains optional dependencies and path input for NewConfigLoader.
type ConfigLoaderOptions struct {
	// Store reads the ordinary user config file; JSONConfigFileStore is used when nil.
	Store ConfigFileStore

	// Resolver resolves concrete config and data paths; DefaultPathResolver is used when nil.
	Resolver PathResolver

	// EnvReader reads environment overrides; os.LookupEnv is used by ReadEnvOverrides when nil.
	EnvReader EnvReader

	// PathInput carries caller-provided app name, mode, and deterministic roots.
	PathInput PathResolveInput
}

// ConfigSourceState reports which source layers contributed to a loaded configuration.
type ConfigSourceState struct {
	// DefaultsApplied reports that DefaultAppConfig seeded the merge.
	DefaultsApplied bool

	// FileState reports whether the ordinary config file was missing or present.
	FileState FileState

	// ConfigFile is the resolved ordinary config file path used for loading.
	ConfigFile string

	// FileLoaded reports that an existing config file was parsed and merged.
	FileLoaded bool

	// EnvOverrideApplied reports that at least one environment override was accepted and merged.
	EnvOverrideApplied bool

	// EnvSources lists the accepted environment variables by affected config field.
	EnvSources []ConfigOverrideSource

	// EnvWriteBack reports whether environment overrides should be persisted; it is always false.
	EnvWriteBack bool

	// ResolvedPaths contains the concrete paths used by the final AppConfig.
	ResolvedPaths ResolvedPaths
}

// LoadResult contains the final composed config and source metadata for callers.
type LoadResult struct {
	// Config is the final AppConfig after defaults, file config, and environment overrides are merged.
	Config AppConfig

	// Source describes the source layers that contributed to Config.
	Source ConfigSourceState
}

// NewConfigLoader builds a ConfigLoader with defaults for omitted dependencies.
func NewConfigLoader(options ConfigLoaderOptions) ConfigLoader {
	if options.Store == nil {
		options.Store = JSONConfigFileStore{}
	}
	if options.Resolver == nil {
		options.Resolver = DefaultPathResolver{}
	}

	return ConfigLoader{
		Store:     options.Store,
		Resolver:  options.Resolver,
		EnvReader: options.EnvReader,
		PathInput: options.PathInput,
	}
}

// Load returns a complete AppConfig composed in the fixed order: defaults, file config, environment overrides.
//
// Missing config files are treated as normal first-start state. Invalid environment overrides, path
// resolution failures, and present-but-unreadable config files are returned as ConfigError values.
func (loader ConfigLoader) Load() (LoadResult, error) {
	loader = NewConfigLoader(ConfigLoaderOptions{
		Store:     loader.Store,
		Resolver:  loader.Resolver,
		EnvReader: loader.EnvReader,
		PathInput: loader.PathInput,
	})

	env, issues := ReadEnvOverrides(loader.EnvReader, loader.envPathResolver())
	if len(issues) != 0 {
		return LoadResult{Source: sourceFromEnv(env)}, ConfigError{
			Code:    primaryIssueCode(issues),
			Message: "环境覆盖无效",
			Issues:  issues,
		}
	}

	readResolved, pathIssues := loader.resolveLoadPaths(env)
	if len(pathIssues) != 0 {
		return LoadResult{Source: sourceFromEnv(env)}, ConfigError{
			Code:    ConfigIssueCodeConfigPathInvalid,
			Message: "配置路径无效",
			Issues:  pathIssues,
		}
	}

	fileConfig, fileState, err := loader.Store.Read(readResolved.ConfigFile)
	if err != nil {
		if configErr, ok := err.(ConfigError); ok {
			return LoadResult{Source: sourceFromEnv(env)}, configErr
		}
		return LoadResult{Source: sourceFromEnv(env)}, ConfigError{
			Code:    ConfigIssueCodeConfigLoadFailed,
			Message: "配置文件读取失败",
			Issues: []ConfigIssue{{
				Path:     "paths.configFile",
				Code:     ConfigIssueCodeConfigLoadFailed,
				Severity: ConfigIssueSeverityError,
				Message:  "配置文件存在但无法读取或解析",
			}},
		}
	}

	cfg := DefaultAppConfig()
	cfg.Paths.ConfigFile = readResolved.ConfigFile
	cfg.Paths.DataDir = readResolved.DataDir
	cfg.Development.Mode = readResolved.Mode
	cfg.Development.UseIsolatedDataDir = readResolved.Isolated

	fileDataDir := fileDataDirOverride(fileConfig, fileState)
	if fileState == FileStatePresent {
		cfg = mergeUserConfig(cfg, fileConfig)
	}
	cfg = applyEnvOverrides(cfg, env)
	finalResolved, finalPathIssues := loader.resolveFinalPaths(cfg, env, fileDataDir)
	if len(finalPathIssues) != 0 {
		return LoadResult{Source: sourceFromEnv(env)}, ConfigError{
			Code:    ConfigIssueCodeConfigPathInvalid,
			Message: "配置路径无效",
			Issues:  finalPathIssues,
		}
	}
	cfg.Paths.ConfigFile = finalResolved.ConfigFile
	cfg.Paths.DataDir = finalResolved.DataDir
	cfg.Development.UseIsolatedDataDir = finalResolved.Isolated
	if validationIssues := (ConfigValidator{}).ValidateForLoad(cfg); len(validationIssues) != 0 {
		return LoadResult{Source: sourceFromEnv(env)}, ConfigError{
			Code:    primaryIssueCode(validationIssues),
			Message: "配置校验失败",
			Issues:  validationIssues,
		}
	}

	return LoadResult{
		Config: cfg,
		Source: ConfigSourceState{
			DefaultsApplied:    true,
			FileState:          fileState,
			ConfigFile:         readResolved.ConfigFile,
			FileLoaded:         fileState == FileStatePresent,
			EnvOverrideApplied: len(env.Sources) != 0,
			EnvSources:         append([]ConfigOverrideSource(nil), env.Sources...),
			EnvWriteBack:       env.WriteBack,
			ResolvedPaths:      finalResolved,
		},
	}, nil
}

// SettingsView converts the loaded AppConfig into the facade-readable settings view.
func (result LoadResult) SettingsView() SettingsView {
	return NewSettingsView(result.Config)
}

func (loader ConfigLoader) envPathResolver() PathResolver {
	return loaderPathResolver{
		base:     loader.PathInput,
		resolver: loader.Resolver,
	}
}

func (loader ConfigLoader) resolveFinalPaths(cfg AppConfig, env EnvOverrides, fileDataDir string) (ResolvedPaths, []ConfigIssue) {
	if env.ResolvedPaths != nil {
		return *env.ResolvedPaths, nil
	}

	input := loader.PathInput
	input.Mode = cfg.Development.Mode
	if fileDataDir != "" {
		input.DataDirOverride = fileDataDir
	}

	return loader.Resolver.Resolve(input)
}

func (loader ConfigLoader) resolveLoadPaths(env EnvOverrides) (ResolvedPaths, []ConfigIssue) {
	if env.ResolvedPaths != nil {
		return *env.ResolvedPaths, nil
	}
	input := mergeEnvPathInput(loader.PathInput, env)
	return loader.Resolver.Resolve(input)
}

// loaderPathResolver preserves caller-provided roots while ReadEnvOverrides validates path override inputs.
type loaderPathResolver struct {
	// base is the loader-level path input supplied by the caller.
	base PathResolveInput

	// resolver is the concrete resolver used to validate merged input.
	resolver PathResolver
}

// Resolve overlays env path input onto the loader-level path input and delegates to the real resolver.
func (resolver loaderPathResolver) Resolve(input PathResolveInput) (ResolvedPaths, []ConfigIssue) {
	return resolver.resolver.Resolve(overlayPathInput(resolver.base, input))
}

func mergeEnvPathInput(base PathResolveInput, env EnvOverrides) PathResolveInput {
	envInput := PathResolveInput{}
	if env.Development != nil && env.Development.Mode != nil {
		envInput.Mode = *env.Development.Mode
	}
	if env.Paths != nil {
		envInput.ConfigDirOverride = env.Paths.ConfigDir
		envInput.DataDirOverride = env.Paths.DataDir
		envInput.TestRoot = env.Paths.TestRoot
		envInput.DevelopmentRoot = env.Paths.DevelopmentRoot
	}
	return overlayPathInput(base, envInput)
}

func overlayPathInput(base PathResolveInput, override PathResolveInput) PathResolveInput {
	if override.AppName != "" {
		base.AppName = override.AppName
	}
	if override.Mode != "" {
		base.Mode = override.Mode
	}
	if override.ConfigDirOverride != "" {
		base.ConfigDirOverride = override.ConfigDirOverride
	}
	if override.DataDirOverride != "" {
		base.DataDirOverride = override.DataDirOverride
	}
	if override.TestRoot != "" {
		base.TestRoot = override.TestRoot
	}
	if override.DevelopmentRoot != "" {
		base.DevelopmentRoot = override.DevelopmentRoot
	}
	if override.DesktopConfigRoot != "" {
		base.DesktopConfigRoot = override.DesktopConfigRoot
	}
	if override.DesktopDataRoot != "" {
		base.DesktopDataRoot = override.DesktopDataRoot
	}
	return base
}

func mergeUserConfig(cfg AppConfig, user UserConfig) AppConfig {
	if user.Version != 0 {
		cfg.Version = user.Version
	}
	if user.Appearance != nil {
		if user.Appearance.Language != "" {
			cfg.Appearance.Language = user.Appearance.Language
		}
		if user.Appearance.Theme != "" {
			cfg.Appearance.Theme = user.Appearance.Theme
		}
	}
	if user.Paths != nil && user.Paths.DataDir != "" {
		cfg.Paths.DataDir = user.Paths.DataDir
	}
	if user.Development != nil {
		if user.Development.Mode != "" {
			cfg.Development.Mode = user.Development.Mode
		}
		if user.Development.UseIsolatedDataDir != nil {
			cfg.Development.UseIsolatedDataDir = *user.Development.UseIsolatedDataDir
		}
		if user.Development.DiagnosticsEnabled != nil {
			cfg.Development.DiagnosticsEnabled = *user.Development.DiagnosticsEnabled
		}
	}
	if user.Integrations != nil {
		cfg.Integrations = mergeFutureIntegrations(cfg.Integrations, *user.Integrations)
	}
	if user.Privacy != nil {
		if user.Privacy.LocalOnly != nil {
			cfg.Privacy.LocalOnly = *user.Privacy.LocalOnly
		}
		if user.Privacy.TelemetryEnabled != nil {
			cfg.Privacy.TelemetryEnabled = *user.Privacy.TelemetryEnabled
		}
	}
	return cfg
}

func fileDataDirOverride(user UserConfig, fileState FileState) string {
	if fileState != FileStatePresent || user.Paths == nil {
		return ""
	}
	return user.Paths.DataDir
}

func mergeFutureIntegrations(current FutureIntegrationsConfig, user FutureIntegrationsConfig) FutureIntegrationsConfig {
	current.Account = mergeFutureIntegration(current.Account, user.Account)
	current.LLM = mergeFutureIntegration(current.LLM, user.LLM)
	return current
}

func mergeFutureIntegration(current FutureIntegrationConfig, user FutureIntegrationConfig) FutureIntegrationConfig {
	current.Enabled = user.Enabled
	current.Configured = user.Configured
	if user.Status != "" {
		current.Status = user.Status
	}
	return current
}

func applyEnvOverrides(cfg AppConfig, env EnvOverrides) AppConfig {
	if env.Appearance != nil {
		if env.Appearance.Language != nil {
			cfg.Appearance.Language = *env.Appearance.Language
		}
		if env.Appearance.Theme != nil {
			cfg.Appearance.Theme = *env.Appearance.Theme
		}
	}
	if env.ResolvedPaths != nil {
		if env.ResolvedPaths.DataDir != "" {
			cfg.Paths.DataDir = env.ResolvedPaths.DataDir
		}
		if env.ResolvedPaths.ConfigFile != "" {
			cfg.Paths.ConfigFile = env.ResolvedPaths.ConfigFile
		}
		cfg.Development.UseIsolatedDataDir = env.ResolvedPaths.Isolated
	}
	if env.Paths != nil && env.ResolvedPaths == nil && env.Paths.DataDir != "" {
		cfg.Paths.DataDir = env.Paths.DataDir
	}
	if env.Development != nil {
		if env.Development.Mode != nil {
			cfg.Development.Mode = *env.Development.Mode
		}
		if env.Development.DiagnosticsEnabled != nil {
			cfg.Development.DiagnosticsEnabled = *env.Development.DiagnosticsEnabled
		}
	}
	return cfg
}

func sourceFromEnv(env EnvOverrides) ConfigSourceState {
	return ConfigSourceState{
		DefaultsApplied:    true,
		EnvOverrideApplied: len(env.Sources) != 0,
		EnvSources:         append([]ConfigOverrideSource(nil), env.Sources...),
		EnvWriteBack:       env.WriteBack,
	}
}

func primaryIssueCode(issues []ConfigIssue) ConfigIssueCode {
	if len(issues) == 0 {
		return ConfigIssueCodeInternalError
	}
	if issues[0].Code == "" {
		return ConfigIssueCodeConfigInvalid
	}
	return issues[0].Code
}
