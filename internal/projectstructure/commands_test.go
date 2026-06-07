package projectstructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskfileDefinesCanonicalCommands(t *testing.T) {
	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "Taskfile.yml"))
	if err != nil {
		t.Fatalf("read Taskfile.yml: %v", err)
	}

	content := string(data)
	for _, command := range []string{"setup", "doctor", "dev", "build", "format", "lint", "test"} {
		if !strings.Contains(content, command+":") {
			t.Fatalf("expected Taskfile.yml to define %q command", command)
		}
	}
}

func TestDoctorScriptContainsActionablePrerequisiteHints(t *testing.T) {
	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "scripts", "doctor.go"))
	if err != nil {
		t.Fatalf("read scripts/doctor.go: %v", err)
	}

	content := string(data)
	for _, expected := range []string{"Go 1.25+", "Node.js", "npm", "wails3", "go install github.com/wailsapp/wails/v3/cmd/wails3@latest"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected doctor.go to mention %q", expected)
		}
	}
}
