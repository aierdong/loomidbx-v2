package main

import "testing"

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
