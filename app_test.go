package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gerdong/loomidbx/internal/config"
)

func TestAppFacadeExposesBootstrapStatus(t *testing.T) {
	app := NewApp()

	status := app.BootstrapStatus()

	if status.Name != ApplicationName {
		t.Fatalf("Name = %q, want %q", status.Name, ApplicationName)
	}
	if status.Version != ApplicationVersion {
		t.Fatalf("Version = %q, want %q", status.Version, ApplicationVersion)
	}
	if status.Runtime != "go" {
		t.Fatalf("Runtime = %q, want go", status.Runtime)
	}
	if !status.Ready {
		t.Fatal("Ready = false, want true")
	}
}

func TestNewAppInitializesConfigService(t *testing.T) {
	app := NewApp()

	if app.config == nil {
		t.Fatal("config service = nil, want initialized service")
	}
}

func TestGetSettingsForwardsServiceView(t *testing.T) {
	want := config.SettingsView{
		Appearance: config.SettingsAppearanceView{
			Language: config.LanguageEn,
			Theme:    config.ThemeDark,
		},
		Paths: config.SettingsPathView{
			DataDir:    `C:\sentinel\data`,
			ConfigFile: `C:\sentinel\config.json`,
		},
		Development: config.SettingsDevelopmentView{
			Mode:               config.ModeTest,
			UseIsolatedDataDir: true,
			DiagnosticsEnabled: true,
		},
		Integrations: config.SettingsIntegrationsView{
			Account: config.SettingsAccountIntegrationView{Status: config.FutureStatusConfigured},
			LLM:     config.SettingsLLMIntegrationView{Configured: true},
		},
		Privacy: config.SettingsPrivacyView{
			LocalOnly:        false,
			TelemetryEnabled: true,
		},
	}
	service := &recordingConfigService{currentView: want}
	app := newAppWithConfigService(service)

	got, err := app.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetSettings() = %+v, want service view %+v", got, want)
	}
	if service.currentCalls != 1 {
		t.Fatalf("Current() calls = %d, want 1", service.currentCalls)
	}
}

func TestUpdateSettingsForwardsInputAndReturnsUpdatedView(t *testing.T) {
	theme := config.ThemeLight
	wantInput := config.UpdateSettingsInput{
		Appearance: &config.UpdateAppearanceInput{Theme: &theme},
	}
	wantView := config.SettingsView{
		Appearance: config.SettingsAppearanceView{
			Language: config.LanguageZh,
			Theme:    config.ThemeLight,
		},
	}
	service := &recordingConfigService{updateView: wantView}
	app := newAppWithConfigService(service)

	got, err := app.UpdateSettings(wantInput)
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, wantView) {
		t.Fatalf("UpdateSettings() = %+v, want service view %+v", got, wantView)
	}
	if !reflect.DeepEqual(service.lastUpdateInput, wantInput) {
		t.Fatalf("UpdateSettings() forwarded input = %+v, want %+v", service.lastUpdateInput, wantInput)
	}
}

func TestUpdateSettingsConvertsFieldIssuesToConfigError(t *testing.T) {
	issues := []config.ConfigIssue{{
		Path:     "appearance.theme",
		Code:     config.ConfigIssueCodeValidationFailed,
		Severity: config.ConfigIssueSeverityError,
		Message:  "主题无效",
	}}
	service := &recordingConfigService{updateIssues: issues}
	app := newAppWithConfigService(service)

	_, err := app.UpdateSettings(config.UpdateSettingsInput{})
	if err == nil {
		t.Fatal("UpdateSettings() error = nil, want ConfigError")
	}
	configErr, ok := err.(config.ConfigError)
	if !ok {
		t.Fatalf("UpdateSettings() error = %T %[1]v, want config.ConfigError", err)
	}
	if configErr.Code != config.ConfigIssueCodeValidationFailed {
		t.Fatalf("ConfigError.Code = %q, want %q", configErr.Code, config.ConfigIssueCodeValidationFailed)
	}
	if !reflect.DeepEqual(configErr.Issues, issues) {
		t.Fatalf("ConfigError.Issues = %+v, want service issues %+v", configErr.Issues, issues)
	}
}

func TestGetSettingsConvertsUnexpectedServiceError(t *testing.T) {
	service := &recordingConfigService{currentErr: errors.New("boom")}
	app := newAppWithConfigService(service)

	_, err := app.GetSettings()
	if err == nil {
		t.Fatal("GetSettings() error = nil, want ConfigError")
	}
	configErr, ok := err.(config.ConfigError)
	if !ok {
		t.Fatalf("GetSettings() error = %T %[1]v, want config.ConfigError", err)
	}
	if configErr.Code != config.ConfigIssueCodeInternalError {
		t.Fatalf("ConfigError.Code = %q, want %q", configErr.Code, config.ConfigIssueCodeInternalError)
	}
}

type recordingConfigService struct {
	currentView config.SettingsView
	currentErr  error

	updateView      config.SettingsView
	updateIssues    []config.ConfigIssue
	updateErr       error
	lastUpdateInput config.UpdateSettingsInput
	currentCalls    int
}

func (service *recordingConfigService) Load() (config.LoadResult, error) {
	return config.LoadResult{}, nil
}

func (service *recordingConfigService) Current() (config.SettingsView, error) {
	service.currentCalls++
	return service.currentView, service.currentErr
}

func (service *recordingConfigService) Update(input config.UpdateSettingsInput) (config.SettingsView, []config.ConfigIssue, error) {
	service.lastUpdateInput = input
	return service.updateView, service.updateIssues, service.updateErr
}
