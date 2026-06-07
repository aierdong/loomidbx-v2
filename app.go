package main

import "github.com/loomidbx/loomidbx-v2/internal/bootstrap"

const (
	ApplicationName    = "LoomiDBX"
	ApplicationVersion = "0.1.0"
)

type App struct {
	bootstrap *bootstrap.Service
}

func NewApp() *App {
	return &App{bootstrap: bootstrap.NewService(ApplicationName, ApplicationVersion)}
}

func (a *App) BootstrapStatus() bootstrap.Status {
	return a.bootstrap.Status()
}
