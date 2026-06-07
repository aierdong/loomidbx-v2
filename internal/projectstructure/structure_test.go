package projectstructure

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDesktopApplicationSkeletonFiles(t *testing.T) {
	root := repositoryRoot(t)

	requiredFiles := []string{
		"go.mod",
		"main.go",
		"app.go",
		"wails.json",
		"frontend/package.json",
		"frontend/index.html",
		"frontend/vite.config.ts",
		"frontend/tsconfig.json",
		"frontend/src/main.ts",
		"frontend/src/App.vue",
		"frontend/generated/README.md",
		"tests/README.md",
		"tests/smoke/README.md",
	}

	for _, name := range requiredFiles {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatalf("expected skeleton file %s to exist: %v", name, err)
		}
	}
}

func TestFrontendPackageDefinesCanonicalScripts(t *testing.T) {
	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "frontend", "package.json"))
	if err != nil {
		t.Fatalf("read frontend package.json: %v", err)
	}

	var pkg struct {
		PackageManager string            `json:"packageManager"`
		Scripts        map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		t.Fatalf("parse frontend package.json: %v", err)
	}

	if pkg.PackageManager != "npm@11" {
		t.Fatalf("expected packageManager npm@11, got %q", pkg.PackageManager)
	}

	for _, script := range []string{"dev", "build", "typecheck", "lint", "format", "test"} {
		if pkg.Scripts[script] == "" {
			t.Fatalf("expected frontend script %q to be defined", script)
		}
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	return filepath.Clean(filepath.Join(cwd, "..", ".."))
}
