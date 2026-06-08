package main

import (
	"errors"
	"testing"
)

func TestMeetsMinimumVersion(t *testing.T) {
	cases := []struct {
		name    string
		version string
		minimum string
		want    bool
	}{
		{name: "exact go version", version: "go version go1.25.0 windows/amd64", minimum: "1.25", want: true},
		{name: "newer go version", version: "go version go1.26.2 windows/amd64", minimum: "1.25", want: true},
		{name: "older go version", version: "go version go1.24.9 windows/amd64", minimum: "1.25", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := meetsMinimumVersion(tc.version, tc.minimum); got != tc.want {
				t.Fatalf("meetsMinimumVersion(%q, %q) = %v, want %v", tc.version, tc.minimum, got, tc.want)
			}
		})
	}
}

func TestEvaluatePrerequisiteReportsActionableMissingTool(t *testing.T) {
	check := prerequisite{
		Name:        "wails3",
		InstallHint: "安装 Wails v3 CLI",
		Required:    true,
	}

	result := evaluatePrerequisite(check, "", errors.New("executable file not found"))

	if result.Status != prerequisiteMissing {
		t.Fatalf("Status = %q, want %q", result.Status, prerequisiteMissing)
	}
	if !result.Blocking {
		t.Fatal("Blocking = false, want true for a missing required tool")
	}
	if result.Action != "安装 Wails v3 CLI" {
		t.Fatalf("Action = %q, want install hint", result.Action)
	}
}

func TestEvaluatePrerequisiteBlocksUnsupportedMinimumVersion(t *testing.T) {
	check := prerequisite{
		Name:        "Go 1.25+",
		Minimum:     "1.25",
		InstallHint: "升级 Go",
		Required:    true,
	}

	result := evaluatePrerequisite(check, "go version go1.24.9 windows/amd64", nil)

	if result.Status != prerequisiteUnsupported {
		t.Fatalf("Status = %q, want %q", result.Status, prerequisiteUnsupported)
	}
	if !result.Blocking {
		t.Fatal("Blocking = false, want true for an unsupported required tool")
	}
	if result.Diagnostic == "" {
		t.Fatal("Diagnostic should explain the minimum version")
	}
}
