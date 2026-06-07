package projectstructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackendPlaceholderModulesAreDocumented(t *testing.T) {
	root := repositoryRoot(t)
	modules := []string{
		"internal/domain/README.md",
		"internal/service/README.md",
		"internal/repository/README.md",
		"internal/config/README.md",
		"internal/storage/README.md",
		"internal/dbx/README.md",
		"internal/dbx/adapter/README.md",
		"internal/dbx/dialect/README.md",
		"internal/dbx/introspect/README.md",
		"internal/dbx/typex/README.md",
		"internal/dbx/capability/README.md",
		"internal/engine/README.md",
		"internal/generator/README.md",
	}

	for _, module := range modules {
		data, err := os.ReadFile(filepath.Join(root, module))
		if err != nil {
			t.Fatalf("expected placeholder documentation %s: %v", module, err)
		}
		content := string(data)
		if !strings.Contains(content, "占位") || !strings.Contains(content, "后续 spec") {
			t.Fatalf("expected %s to describe placeholder and later spec ownership", module)
		}
	}
}
