package main

const (
	ApplicationName    = "LoomiDBX"
	ApplicationVersion = "0.1.0"
)

type App struct {
	name    string
	version string
}

func NewApp() *App {
	return &App{name: ApplicationName, version: ApplicationVersion}
}

func (a *App) Name() string {
	return a.name
}

func (a *App) Version() string {
	return a.version
}
