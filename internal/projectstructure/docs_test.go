package projectstructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectStructureDocumentationExists(t *testing.T) {
	root := repositoryRoot(t)
	checks := map[string][]string{
		"README.md":                              {"工程骨架", "后续 spec", "隐私"},
		"docs/architecture/project-structure.md": {"后端模块", "前端模块", "生成绑定", "phase-01-config-system", "phase-01-local-storage-strategy", "phase-01-database-dialect-interface"},
	}
	for file, expectedTerms := range checks {
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		content := string(data)
		for _, term := range expectedTerms {
			if !strings.Contains(content, term) {
				t.Fatalf("expected %s to mention %q", file, term)
			}
		}
	}
}
