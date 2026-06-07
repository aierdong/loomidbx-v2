package projectstructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandsDocumentationStatesValidationAndPrivacy(t *testing.T) {
	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "docs", "development", "commands.md"))
	if err != nil {
		t.Fatalf("read commands docs: %v", err)
	}
	content := string(data)
	for _, term := range []string{"setup", "doctor", "dev", "build", "format", "lint", "test", "deferred", "phase-01-test-tooling", "数据库凭据", "用户 SQL"} {
		if !strings.Contains(content, term) {
			t.Fatalf("expected commands docs to mention %q", term)
		}
	}
}
