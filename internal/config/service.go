package config

// Service defines the backend configuration entrypoint independent of Wails UI bindings.
type Service interface {
	// Load returns the composed, validated application config and source metadata.
	//
	// Callers receive ConfigError when environment overrides, paths, file loading, or merged config
	// validation fail.
	Load() (LoadResult, error)

	// Current returns the settings view derived from the composed, validated application config.
	//
	// Callers receive ConfigError when the underlying load operation cannot produce a valid config.
	Current() (SettingsView, error)

	// Update validates and persists user-editable settings, then returns the reloaded settings view.
	//
	// Validation failures are returned as ConfigIssue values with a nil error and no file write. Write or
	// reload failures are returned as ConfigError values.
	Update(input UpdateSettingsInput) (SettingsView, []ConfigIssue, error)
}

// ConfigService implements the backend configuration read and update use cases.
type ConfigService struct {
	// loader composes defaults, file config, and environment overrides for every service operation.
	loader ConfigLoader
}

// ConfigServiceOptions contains dependencies used to build a ConfigService.
type ConfigServiceOptions struct {
	// Loader is the config loader used by service operations; default dependencies are filled when omitted.
	Loader ConfigLoader
}

// NewConfigService builds a ConfigService with default loader dependencies when omitted.
func NewConfigService(options ConfigServiceOptions) ConfigService {
	return ConfigService{
		loader: normalizeLoader(options.Loader),
	}
}

// Load returns the composed, validated application config and source metadata without depending on Wails UI.
func (service ConfigService) Load() (LoadResult, error) {
	return service.normalizedLoader().Load()
}

// Current returns the facade-readable settings view for the current composed and validated config.
func (service ConfigService) Current() (SettingsView, error) {
	result, err := service.Load()
	if err != nil {
		return SettingsView{}, err
	}
	return result.SettingsView(), nil
}

// Update validates and saves user-editable settings, then reloads and returns the resulting settings view.
func (service ConfigService) Update(input UpdateSettingsInput) (SettingsView, []ConfigIssue, error) {
	loader := service.normalizedLoader()
	loaded, err := loader.Load()
	if err != nil {
		return SettingsView{}, nil, err
	}

	targetPath := loaded.Source.ConfigFile
	if targetPath == "" {
		targetPath = loaded.Config.Paths.ConfigFile
	}

	existing, state, err := loader.Store.Read(targetPath)
	if err != nil {
		return SettingsView{}, nil, ConfigError{
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
	if state == FileStateMissing {
		existing = UserConfig{Version: CurrentConfigVersion}
	}

	candidate := applySettingsUpdate(existing, input)
	if candidate.Version == 0 {
		candidate.Version = CurrentConfigVersion
	}

	if issues := validateServiceCandidate(loader, candidate); len(issues) != 0 {
		return SettingsView{}, issues, nil
	}

	if err := loader.Store.Save(targetPath, candidate); err != nil {
		if configErr, ok := err.(ConfigError); ok {
			if configErr.Code == ConfigIssueCodeValidationFailed || configErr.Code == ConfigIssueCodeConfigInvalid || configErr.Code == ConfigIssueCodeConfigPathInvalid {
				return SettingsView{}, configErr.Issues, nil
			}
		}
		return SettingsView{}, nil, writeFailureError(err)
	}

	reloaded, err := loader.Load()
	if err != nil {
		return SettingsView{}, nil, err
	}
	return reloaded.SettingsView(), nil, nil
}

func normalizeLoader(loader ConfigLoader) ConfigLoader {
	return NewConfigLoader(ConfigLoaderOptions{
		Store:     loader.Store,
		Resolver:  loader.Resolver,
		EnvReader: loader.EnvReader,
		PathInput: loader.PathInput,
	})
}

func (service ConfigService) normalizedLoader() ConfigLoader {
	return normalizeLoader(service.loader)
}

func applySettingsUpdate(config UserConfig, input UpdateSettingsInput) UserConfig {
	if input.Appearance != nil {
		if config.Appearance == nil {
			config.Appearance = &AppearanceConfig{}
		}
		if input.Appearance.Language != nil {
			config.Appearance.Language = *input.Appearance.Language
		}
		if input.Appearance.Theme != nil {
			config.Appearance.Theme = *input.Appearance.Theme
		}
	}

	if input.Paths != nil {
		if config.Paths == nil {
			config.Paths = &UserPathConfig{}
		}
		if input.Paths.DataDir != nil {
			config.Paths.DataDir = *input.Paths.DataDir
		}
	}

	if input.Development != nil {
		if config.Development == nil {
			config.Development = &UserDevelopmentConfig{}
		}
		if input.Development.Mode != nil {
			config.Development.Mode = *input.Development.Mode
		}
		if input.Development.UseIsolatedDataDir != nil {
			config.Development.UseIsolatedDataDir = boolPtr(*input.Development.UseIsolatedDataDir)
		}
		if input.Development.DiagnosticsEnabled != nil {
			config.Development.DiagnosticsEnabled = boolPtr(*input.Development.DiagnosticsEnabled)
		}
	}

	if input.Privacy != nil {
		if config.Privacy == nil {
			config.Privacy = &UserPrivacyConfig{}
		}
		if input.Privacy.LocalOnly != nil {
			config.Privacy.LocalOnly = boolPtr(*input.Privacy.LocalOnly)
		}
		if input.Privacy.TelemetryEnabled != nil {
			config.Privacy.TelemetryEnabled = boolPtr(*input.Privacy.TelemetryEnabled)
		}
	}

	return config
}

func validateServiceCandidate(loader ConfigLoader, config UserConfig) []ConfigIssue {
	pathInput := loader.PathInput
	pathInput.Mode = ModeDesktop
	pathInput.ConfigDirOverride = ""
	pathInput.DataDirOverride = ""
	validationLoader := normalizeLoader(ConfigLoader{
		Store:     fixedConfigStore{config: config, state: FileStatePresent},
		Resolver:  loader.Resolver,
		EnvReader: mapEnvReader(nil),
		PathInput: pathInput,
	})
	_, err := validationLoader.Load()
	if err == nil {
		return nil
	}
	if configErr, ok := err.(ConfigError); ok {
		return configErr.Issues
	}
	return []ConfigIssue{{
		Path:     "config",
		Code:     ConfigIssueCodeInternalError,
		Severity: ConfigIssueSeverityError,
		Message:  "配置校验失败",
	}}
}

func writeFailureError(err error) ConfigError {
	message := "配置文件写入失败"
	return ConfigError{
		Code:    ConfigIssueCodeConfigWriteFailed,
		Message: message,
		Issues: []ConfigIssue{{
			Path:     "paths.configFile",
			Code:     ConfigIssueCodeConfigWriteFailed,
			Severity: ConfigIssueSeverityError,
			Message:  message,
		}},
	}
}

// fixedConfigStore provides an in-memory read-only config file for service candidate validation.
type fixedConfigStore struct {
	// config is the user config returned from Read regardless of path.
	config UserConfig

	// state is the file state returned from Read regardless of path.
	state FileState
}

func (store fixedConfigStore) Read(path string) (UserConfig, FileState, error) {
	return store.config, store.state, nil
}

func (store fixedConfigStore) Save(path string, config UserConfig) error {
	return nil
}
