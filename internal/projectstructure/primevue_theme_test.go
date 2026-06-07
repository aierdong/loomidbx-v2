package projectstructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrimeVueThemeAndStylesAreDocumented(t *testing.T) {
	root := repositoryRoot(t)
	files := []string{
		"frontend/src/styles/global.css",
		"frontend/src/styles/tokens.css",
		"frontend/src/styles/utilities.css",
		"frontend/src/styles/README.md",
		"frontend/src/ui/primevue/preset.ts",
		"frontend/src/ui/primevue/config.ts",
		"frontend/src/ui/primevue/README.md",
	}
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatalf("expected PrimeVue/style file %s: %v", file, err)
		}
		if strings.TrimSpace(string(data)) == "" {
			t.Fatalf("expected %s to have content", file)
		}
	}
}
