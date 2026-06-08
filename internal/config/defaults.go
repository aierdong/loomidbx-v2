package config

const (
	DefaultDataDirPlaceholder    = "{appDataDir}"
	DefaultConfigFilePlaceholder = "{appConfigDir}/config.json"
	PlaintextCredentialsDenied   = "plaintext credentials are not stored in ordinary app config"
)

func DefaultAppConfig() AppConfig {
	return AppConfig{
		Version: CurrentConfigVersion,
		Appearance: AppearanceConfig{
			Language: LanguageZh,
			Theme:    ThemeSystem,
		},
		Paths: PathConfig{
			DataDir:    DefaultDataDirPlaceholder,
			ConfigFile: DefaultConfigFilePlaceholder,
		},
		Development: DevelopmentConfig{
			Mode:               ModeDesktop,
			UseIsolatedDataDir: false,
			DiagnosticsEnabled: false,
		},
		Integrations: FutureIntegrationsConfig{
			Account: defaultFutureIntegration(),
			LLM:     defaultFutureIntegration(),
		},
		Privacy: PrivacyConfig{
			LocalOnly:        true,
			TelemetryEnabled: false,
			BusinessData: BusinessDataBoundary{
				StoredInAppConfig: false,
			},
			SensitiveCredentials: SensitiveCredentialsBoundary{
				StoredInAppConfig:       false,
				ExternalStorageRequired: true,
				PlaintextPolicy:         PlaintextCredentialsDenied,
			},
			NetworkUploadDisabled: true,
		},
	}
}

func DefaultUserConfig() UserConfig {
	cfg := DefaultAppConfig()

	return UserConfig{
		Version:    cfg.Version,
		Appearance: &cfg.Appearance,
		Paths: &UserPathConfig{
			DataDir: cfg.Paths.DataDir,
		},
		Development: &UserDevelopmentConfig{
			Mode:               cfg.Development.Mode,
			UseIsolatedDataDir: cfg.Development.UseIsolatedDataDir,
			DiagnosticsEnabled: cfg.Development.DiagnosticsEnabled,
		},
		Integrations: &cfg.Integrations,
		Privacy: &UserPrivacyConfig{
			LocalOnly:        cfg.Privacy.LocalOnly,
			TelemetryEnabled: cfg.Privacy.TelemetryEnabled,
		},
	}
}

func defaultFutureIntegration() FutureIntegrationConfig {
	return FutureIntegrationConfig{
		Enabled:    false,
		Configured: false,
		Status:     FutureStatusUnavailable,
	}
}
