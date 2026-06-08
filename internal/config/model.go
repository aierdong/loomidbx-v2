package config

const CurrentConfigVersion = 1

type Language string

const (
	LanguageZh Language = "zh"
	LanguageEn Language = "en"
)

type Theme string

const (
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
	ThemeSystem Theme = "system"
)

type RunMode string

const (
	ModeDesktop     RunMode = "desktop"
	ModeDevelopment RunMode = "development"
	ModeTest        RunMode = "test"
)

type FutureIntegrationStatus string

const (
	FutureStatusUnavailable   FutureIntegrationStatus = "unavailable"
	FutureStatusNotConfigured FutureIntegrationStatus = "not_configured"
	FutureStatusConfigured    FutureIntegrationStatus = "configured"
)

type AppConfig struct {
	Version      int                      `json:"version"`
	Appearance   AppearanceConfig         `json:"appearance"`
	Paths        PathConfig               `json:"paths"`
	Development  DevelopmentConfig        `json:"development"`
	Integrations FutureIntegrationsConfig `json:"integrations"`
	Privacy      PrivacyConfig            `json:"privacy"`
}

type AppearanceConfig struct {
	Language Language `json:"language"`
	Theme    Theme    `json:"theme"`
}

type PathConfig struct {
	DataDir    string `json:"dataDir"`
	ConfigFile string `json:"configFile"`
}

type DevelopmentConfig struct {
	Mode               RunMode `json:"mode"`
	UseIsolatedDataDir bool    `json:"useIsolatedDataDir"`
	DiagnosticsEnabled bool    `json:"diagnosticsEnabled"`
}

type FutureIntegrationsConfig struct {
	Account FutureIntegrationConfig `json:"account"`
	LLM     FutureIntegrationConfig `json:"llm"`
}

type FutureIntegrationConfig struct {
	Enabled    bool                    `json:"enabled"`
	Configured bool                    `json:"configured"`
	Status     FutureIntegrationStatus `json:"status"`
}

type PrivacyConfig struct {
	LocalOnly             bool                         `json:"localOnly"`
	TelemetryEnabled      bool                         `json:"telemetryEnabled"`
	BusinessData          BusinessDataBoundary         `json:"businessData"`
	SensitiveCredentials  SensitiveCredentialsBoundary `json:"sensitiveCredentials"`
	NetworkUploadDisabled bool                         `json:"networkUploadDisabled"`
}

type BusinessDataBoundary struct {
	StoredInAppConfig bool `json:"storedInAppConfig"`
}

type SensitiveCredentialsBoundary struct {
	StoredInAppConfig       bool   `json:"storedInAppConfig"`
	ExternalStorageRequired bool   `json:"externalStorageRequired"`
	PlaintextPolicy         string `json:"plaintextPolicy"`
}

type UserConfig struct {
	Version      int                       `json:"version"`
	Appearance   *AppearanceConfig         `json:"appearance,omitempty"`
	Paths        *UserPathConfig           `json:"paths,omitempty"`
	Development  *UserDevelopmentConfig    `json:"development,omitempty"`
	Integrations *FutureIntegrationsConfig `json:"integrations,omitempty"`
	Privacy      *UserPrivacyConfig        `json:"privacy,omitempty"`
}

type UserPathConfig struct {
	DataDir string `json:"dataDir,omitempty"`
}

type UserDevelopmentConfig struct {
	Mode               RunMode `json:"mode,omitempty"`
	UseIsolatedDataDir bool    `json:"useIsolatedDataDir,omitempty"`
	DiagnosticsEnabled bool    `json:"diagnosticsEnabled,omitempty"`
}

type UserPrivacyConfig struct {
	LocalOnly        bool `json:"localOnly"`
	TelemetryEnabled bool `json:"telemetryEnabled"`
}
