package projectstructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontendApiClientBoundaryFiles(t *testing.T) {
	root := repositoryRoot(t)
	files := []string{
		"frontend/src/api/bootstrapClient.ts",
		"frontend/src/api/result.ts",
		"frontend/src/api/README.md",
		"frontend/src/types/bootstrap.ts",
		"frontend/src/types/README.md",
	}

	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatalf("expected frontend api boundary file %s: %v", file, err)
		}
		if !strings.Contains(string(data), "generated") && strings.HasSuffix(file, "api/README.md") {
			t.Fatalf("expected %s to document generated binding boundary", file)
		}
	}
}
