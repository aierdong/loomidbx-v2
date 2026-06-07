package projectstructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontendShellRouterAndStoreFiles(t *testing.T) {
	root := repositoryRoot(t)
	files := []string{
		"frontend/src/pages/BootstrapPage.vue",
		"frontend/src/pages/README.md",
		"frontend/src/components/AppStatusCard.vue",
		"frontend/src/components/README.md",
		"frontend/src/stores/bootstrapStore.ts",
		"frontend/src/stores/README.md",
		"frontend/src/router/index.ts",
		"frontend/src/router/README.md",
	}

	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatalf("expected frontend shell file %s: %v", file, err)
		}
		if strings.TrimSpace(string(data)) == "" {
			t.Fatalf("expected %s to have content", file)
		}
	}
}
