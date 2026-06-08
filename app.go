package main

import (
	"errors"

	"github.com/gerdong/loomidbx/internal/bootstrap"
	"github.com/gerdong/loomidbx/internal/config"
)

const (
	// ApplicationName is the desktop application name used by bootstrap and config path resolution.
	ApplicationName = "LoomiDBX"

	// ApplicationVersion is the current desktop application version exposed by bootstrap status.
	ApplicationVersion = "0.1.0"
)

// App is the Wails-facing application facade composed from backend services.
type App struct {
	// bootstrap exposes startup status for the desktop shell.
	bootstrap *bootstrap.Service

	// config exposes the backend configuration service through thin facade methods.
	config config.Service
}

// NewApp initializes the application composition root with backend services used by Wails bindings.
func NewApp() *App {
	return newAppWithConfigService(config.NewConfigService(config.ConfigServiceOptions{}))
}

// BootstrapStatus returns the current bootstrap status for the desktop shell.
func (a *App) BootstrapStatus() bootstrap.Status {
	return a.bootstrap.Status()
}

// GetSettings forwards settings reads to the backend configuration service and returns facade-safe errors.
func (a *App) GetSettings() (config.SettingsView, error) {
	view, err := a.config.Current()
	if err != nil {
		return config.SettingsView{}, facadeConfigError(config.ConfigIssueCodeInternalError, "设置读取失败", err, nil)
	}
	return view, nil
}

// UpdateSettings forwards settings updates to the backend configuration service.
//
// Validation issues are converted into ConfigError so Go and Wails callers receive field-level details.
func (a *App) UpdateSettings(input config.UpdateSettingsInput) (config.SettingsView, error) {
	view, issues, err := a.config.Update(input)
	if err != nil {
		return config.SettingsView{}, facadeConfigError(config.ConfigIssueCodeInternalError, "设置更新失败", err, nil)
	}
	if len(issues) != 0 {
		return config.SettingsView{}, facadeConfigError(primaryFacadeIssueCode(issues), "设置校验失败", nil, issues)
	}
	return view, nil
}

func newAppWithConfigService(configService config.Service) *App {
	return &App{
		bootstrap: bootstrap.NewService(ApplicationName, ApplicationVersion),
		config:    configService,
	}
}

func facadeConfigError(defaultCode config.ConfigIssueCode, defaultMessage string, err error, issues []config.ConfigIssue) error {
	if err != nil {
		var configErr config.ConfigError
		if errors.As(err, &configErr) {
			return configErr
		}
	}
	return config.ConfigError{
		Code:    defaultCode,
		Message: defaultMessage,
		Issues:  append([]config.ConfigIssue(nil), issues...),
	}
}

func primaryFacadeIssueCode(issues []config.ConfigIssue) config.ConfigIssueCode {
	if len(issues) == 0 || issues[0].Code == "" {
		return config.ConfigIssueCodeInternalError
	}
	return issues[0].Code
}
