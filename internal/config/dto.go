package config

// ConfigIssueCode identifies the stable error category returned by config loading, validation, and saving.
type ConfigIssueCode string

const (
	// ConfigIssueCodeConfigInvalid reports an already-merged configuration that cannot be accepted.
	ConfigIssueCodeConfigInvalid ConfigIssueCode = "CONFIG_INVALID"

	// ConfigIssueCodeConfigLoadFailed reports a configuration file read or parse failure.
	ConfigIssueCodeConfigLoadFailed ConfigIssueCode = "CONFIG_LOAD_FAILED"

	// ConfigIssueCodeValidationFailed reports an invalid field value.
	ConfigIssueCodeValidationFailed ConfigIssueCode = "VALIDATION_FAILED"

	// ConfigIssueCodeConfigPathInvalid reports a path that cannot be used safely.
	ConfigIssueCodeConfigPathInvalid ConfigIssueCode = "CONFIG_PATH_INVALID"

	// ConfigIssueCodeSensitiveValueNotAllowed reports sensitive plaintext in ordinary config.
	ConfigIssueCodeSensitiveValueNotAllowed ConfigIssueCode = "SENSITIVE_VALUE_NOT_ALLOWED"

	// ConfigIssueCodeConfigWriteFailed reports a failed config file save.
	ConfigIssueCodeConfigWriteFailed ConfigIssueCode = "CONFIG_WRITE_FAILED"

	// ConfigIssueCodeInternalError reports an unexpected internal configuration error.
	ConfigIssueCodeInternalError ConfigIssueCode = "INTERNAL_ERROR"
)

// ConfigIssueSeverity classifies how callers should present or handle a config issue.
type ConfigIssueSeverity string

const (
	// ConfigIssueSeverityInfo marks a non-blocking informational issue.
	ConfigIssueSeverityInfo ConfigIssueSeverity = "info"

	// ConfigIssueSeverityWarning marks a warning that callers may show without blocking startup.
	ConfigIssueSeverityWarning ConfigIssueSeverity = "warning"

	// ConfigIssueSeverityError marks an issue that blocks accepting or saving config.
	ConfigIssueSeverityError ConfigIssueSeverity = "error"
)

// ConfigIssue is a field-level config problem reused by load, validation, and save flows.
type ConfigIssue struct {
	// Path is the dotted config field path, such as appearance.theme.
	Path string `json:"path"`

	// Code is the stable machine-readable issue category.
	Code ConfigIssueCode `json:"code"`

	// Severity tells callers whether the issue is informational, warning, or blocking.
	Severity ConfigIssueSeverity `json:"severity"`

	// Message is a user-readable explanation that must not contain sensitive values.
	Message string `json:"message"`
}

// ConfigError is the facade-friendly error envelope for configuration failures.
type ConfigError struct {
	// Code is the primary error category for the failed operation.
	Code ConfigIssueCode `json:"code"`

	// Message is the user-readable summary for the failed operation.
	Message string `json:"message"`

	// Issues carries field-level details when validation or path checks fail.
	Issues []ConfigIssue `json:"issues"`
}

// Error returns the summary message so ConfigError can be used as a Go error.
func (err ConfigError) Error() string {
	return err.Message
}

// SettingsView is the resolved, validated settings shape exposed to backend callers and the Wails facade.
type SettingsView struct {
	// Appearance contains user-facing language and theme settings.
	Appearance SettingsAppearanceView `json:"appearance"`

	// Paths contains resolved local config and data locations.
	Paths SettingsPathView `json:"paths"`

	// Development contains current runtime mode and development-only toggles.
	Development SettingsDevelopmentView `json:"development"`

	// Integrations contains non-sensitive status for future account and LLM entries.
	Integrations SettingsIntegrationsView `json:"integrations"`

	// Privacy contains local-first privacy flags visible to the settings contract.
	Privacy SettingsPrivacyView `json:"privacy"`
}

// SettingsAppearanceView is the frontend-readable appearance section.
type SettingsAppearanceView struct {
	// Language is the active UI language.
	Language Language `json:"language"`

	// Theme is the active UI theme mode.
	Theme Theme `json:"theme"`
}

// SettingsPathView is the frontend-readable resolved path section.
type SettingsPathView struct {
	// DataDir is the resolved local data directory.
	DataDir string `json:"dataDir"`

	// ConfigFile is the resolved ordinary config file path.
	ConfigFile string `json:"configFile"`
}

// SettingsDevelopmentView is the frontend-readable runtime mode section.
type SettingsDevelopmentView struct {
	// Mode is the current desktop, development, or test mode.
	Mode RunMode `json:"mode"`

	// UseIsolatedDataDir indicates whether development or test isolation is active.
	UseIsolatedDataDir bool `json:"useIsolatedDataDir"`

	// DiagnosticsEnabled indicates whether local diagnostics are enabled.
	DiagnosticsEnabled bool `json:"diagnosticsEnabled"`
}

// SettingsIntegrationsView is the frontend-readable future integration status section.
type SettingsIntegrationsView struct {
	// Account exposes only the future account entry status.
	Account SettingsAccountIntegrationView `json:"account"`

	// LLM exposes only non-sensitive LLM configuration state.
	LLM SettingsLLMIntegrationView `json:"llm"`
}

// SettingsAccountIntegrationView is the future account entry status visible to settings.
type SettingsAccountIntegrationView struct {
	// Status describes whether the future account entry is unavailable, configured, or not configured.
	Status FutureIntegrationStatus `json:"status"`
}

// SettingsLLMIntegrationView is the future LLM entry status visible to settings.
type SettingsLLMIntegrationView struct {
	// Configured reports whether a future secure storage boundary has LLM configuration.
	Configured bool `json:"configured"`
}

// SettingsPrivacyView is the frontend-readable privacy section.
type SettingsPrivacyView struct {
	// LocalOnly reports whether the app is operating under local-only privacy semantics.
	LocalOnly bool `json:"localOnly"`

	// TelemetryEnabled reports whether non-sensitive telemetry is enabled.
	TelemetryEnabled bool `json:"telemetryEnabled"`
}

// NewSettingsView converts a resolved AppConfig into the facade settings DTO without exposing sensitive values.
func NewSettingsView(cfg AppConfig) SettingsView {
	return SettingsView{
		Appearance: SettingsAppearanceView{
			Language: cfg.Appearance.Language,
			Theme:    cfg.Appearance.Theme,
		},
		Paths: SettingsPathView{
			DataDir:    cfg.Paths.DataDir,
			ConfigFile: cfg.Paths.ConfigFile,
		},
		Development: SettingsDevelopmentView{
			Mode:               cfg.Development.Mode,
			UseIsolatedDataDir: cfg.Development.UseIsolatedDataDir,
			DiagnosticsEnabled: cfg.Development.DiagnosticsEnabled,
		},
		Integrations: SettingsIntegrationsView{
			Account: SettingsAccountIntegrationView{
				Status: cfg.Integrations.Account.Status,
			},
			LLM: SettingsLLMIntegrationView{
				Configured: cfg.Integrations.LLM.Configured,
			},
		},
		Privacy: SettingsPrivacyView{
			LocalOnly:        cfg.Privacy.LocalOnly,
			TelemetryEnabled: cfg.Privacy.TelemetryEnabled,
		},
	}
}

// UpdateSettingsInput contains the user-editable settings fields accepted by the config service.
type UpdateSettingsInput struct {
	// Appearance optionally updates language or theme.
	Appearance *UpdateAppearanceInput `json:"appearance,omitempty"`

	// Paths optionally updates user-editable path settings.
	Paths *UpdatePathInput `json:"paths,omitempty"`

	// Development optionally updates mode and development toggles.
	Development *UpdateDevelopmentInput `json:"development,omitempty"`

	// Privacy optionally updates privacy flags.
	Privacy *UpdatePrivacyInput `json:"privacy,omitempty"`
}

// UpdateAppearanceInput contains optional appearance updates.
type UpdateAppearanceInput struct {
	// Language updates the UI language when present.
	Language *Language `json:"language,omitempty"`

	// Theme updates the UI theme when present.
	Theme *Theme `json:"theme,omitempty"`
}

// UpdatePathInput contains optional path updates.
type UpdatePathInput struct {
	// DataDir updates the local data directory when present.
	DataDir *string `json:"dataDir,omitempty"`
}

// UpdateDevelopmentInput contains optional runtime mode and diagnostics updates.
type UpdateDevelopmentInput struct {
	// Mode updates the runtime mode when present.
	Mode *RunMode `json:"mode,omitempty"`

	// UseIsolatedDataDir updates development or test data isolation when present.
	UseIsolatedDataDir *bool `json:"useIsolatedDataDir,omitempty"`

	// DiagnosticsEnabled updates local diagnostics when present.
	DiagnosticsEnabled *bool `json:"diagnosticsEnabled,omitempty"`
}

// UpdatePrivacyInput contains optional privacy updates.
type UpdatePrivacyInput struct {
	// LocalOnly updates the local-only privacy flag when present.
	LocalOnly *bool `json:"localOnly,omitempty"`

	// TelemetryEnabled updates the non-sensitive telemetry flag when present.
	TelemetryEnabled *bool `json:"telemetryEnabled,omitempty"`
}
