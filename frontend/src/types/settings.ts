export type Language = 'zh' | 'en'

export type Theme = 'light' | 'dark' | 'system'

export type RunMode = 'desktop' | 'development' | 'test'

export type FutureIntegrationStatus =
  | 'unavailable'
  | 'not_configured'
  | 'configured'

export type ConfigIssueCode =
  | 'CONFIG_INVALID'
  | 'CONFIG_LOAD_FAILED'
  | 'VALIDATION_FAILED'
  | 'CONFIG_PATH_INVALID'
  | 'SENSITIVE_VALUE_NOT_ALLOWED'
  | 'CONFIG_WRITE_FAILED'
  | 'INTERNAL_ERROR'

export type ConfigIssueSeverity = 'info' | 'warning' | 'error'

export interface ConfigIssue {
  path: string
  code: ConfigIssueCode
  severity: ConfigIssueSeverity
  message: string
}

export interface ConfigError {
  code: ConfigIssueCode
  message: string
  issues: ConfigIssue[]
}

export interface SettingsView {
  appearance: SettingsAppearanceView
  paths: SettingsPathView
  development: SettingsDevelopmentView
  integrations: SettingsIntegrationsView
  privacy: SettingsPrivacyView
}

export interface SettingsAppearanceView {
  language: Language
  theme: Theme
}

export interface SettingsPathView {
  dataDir: string
  configFile: string
}

export interface SettingsDevelopmentView {
  mode: RunMode
  useIsolatedDataDir: boolean
  diagnosticsEnabled: boolean
}

export interface SettingsIntegrationsView {
  account: SettingsAccountIntegrationView
  llm: SettingsLLMIntegrationView
}

export interface SettingsAccountIntegrationView {
  status: FutureIntegrationStatus
}

export interface SettingsLLMIntegrationView {
  configured: boolean
}

export interface SettingsPrivacyView {
  localOnly: boolean
  telemetryEnabled: boolean
}

export interface UpdateSettingsInput {
  appearance?: UpdateAppearanceInput
  paths?: UpdatePathInput
  development?: UpdateDevelopmentInput
  privacy?: UpdatePrivacyInput
}

export interface UpdateAppearanceInput {
  language?: Language
  theme?: Theme
}

export interface UpdatePathInput {
  dataDir?: string
}

export interface UpdateDevelopmentInput {
  mode?: RunMode
  useIsolatedDataDir?: boolean
  diagnosticsEnabled?: boolean
}

export interface UpdatePrivacyInput {
  localOnly?: boolean
  telemetryEnabled?: boolean
}
