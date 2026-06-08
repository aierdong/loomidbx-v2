package projectstructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTestToolingTaskfileDefinesLayeredAndAggregateValidation(t *testing.T) {
	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "Taskfile.yml"))
	if err != nil {
		t.Fatalf("read Taskfile.yml: %v", err)
	}

	content := string(data)
	for _, command := range []string{
		"format:go",
		"format:frontend",
		"lint:go",
		"lint:frontend",
		"test:go",
		"test:frontend",
		"build:frontend",
		"build:fallback",
		"verify",
	} {
		if !strings.Contains(content, command+":") {
			t.Fatalf("expected Taskfile.yml to define %q for phase-01-test-tooling", command)
		}
	}

	for _, expected := range []string{
		"task: doctor",
		"task: format",
		"task: lint",
		"task: test",
		"task: build",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected verify to include deterministic stage %q", expected)
		}
	}
}

func TestTestToolingTaskfileFormatsTrackedGoSources(t *testing.T) {
	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "Taskfile.yml"))
	if err != nil {
		t.Fatalf("read Taskfile.yml: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "git ls-files '*.go'") {
		t.Fatalf("expected Go format command to discover tracked Go sources with git ls-files")
	}
	if strings.Contains(content, "internal/projectstructure/*.go") {
		t.Fatalf("expected Go format command not to enumerate only existing skeleton packages")
	}
}

func TestTestToolingDocumentationDefinesReusableValidationMatrix(t *testing.T) {
	root := repositoryRoot(t)
	checks := map[string][]string{
		"README.md": {
			"task verify",
			"docs/development/commands.md",
		},
		filepath.Join("docs", "development", "commands.md"): {
			"后续 spec 引用方式",
			"普通改动必跑",
			"环境依赖",
			"fallback 不能替代完整 Wails build",
			"Vitest",
		},
		filepath.Join("tests", "README.md"): {
			"Go 单元测试",
			"前端 deterministic 样例验证",
			"真实数据库集成",
			"UI E2E",
		},
		filepath.Join("tests", "smoke", "phase-01-validation.md"): {
			"task verify",
			"task build:fallback",
			"不读取真实数据库凭据",
		},
	}

	for file, terms := range checks {
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		content := string(data)
		for _, term := range terms {
			if !strings.Contains(content, term) {
				t.Fatalf("expected %s to mention %q", file, term)
			}
		}
	}
}
