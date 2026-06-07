package projectstructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBootstrapCallChainFilesConnectBoundaries(t *testing.T) {
	root := repositoryRoot(t)
	checks := map[string]string{
		"frontend/generated/bootstrap.ts":       "BootstrapStatus",
		"frontend/src/api/bootstrapClient.ts":   "generated",
		"frontend/src/stores/bootstrapStore.ts": "getStatus",
		"frontend/src/pages/BootstrapPage.vue":  "loadBootstrapStatus",
	}
	for file, expected := range checks {
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if !strings.Contains(string(data), expected) {
			t.Fatalf("expected %s to contain %q", file, expected)
		}
	}
}
