package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigScopeDoesNotAbsorbAdjacentSpecs(t *testing.T) {
	root := findRepoRoot(t)

	productionFiles := []string{
		"app.go",
		"internal/config/defaults.go",
		"internal/config/dto.go",
		"internal/config/env.go",
		"internal/config/loader.go",
		"internal/config/model.go",
		"internal/config/paths.go",
		"internal/config/service.go",
		"internal/config/store.go",
		"internal/config/validate.go",
		"frontend/src/api/settingsClient.ts",
		"frontend/src/types/settings.ts",
	}

	forbidden := map[string]string{
		"net" + "/http":       "配置系统不得引入 HTTP 上传或远端服务依赖",
		"http" + ".Client":    "配置系统不得创建网络客户端",
		"database" + "/sql":   "配置系统不得访问目标数据库或本地数据库连接",
		"CREATE" + " TABLE":   "本地数据库结构创建属于相邻 spec",
		"fetch" + "(":         "前端设置契约不得直接发起网络请求",
		"ax" + "ios":          "前端设置契约不得引入网络 client",
		"XMLHttp" + "Request": "前端设置契约不得直接发起网络请求",
		"Settings" + "Page":   "完整设置页不属于当前 spec",
		"settings" + " page":  "完整设置页不属于当前 spec",
	}

	for _, relativePath := range productionFiles {
		content := readRepoFile(t, root, relativePath)
		for token, reason := range forbidden {
			if strings.Contains(content, token) {
				t.Fatalf("%s contains forbidden token %q: %s", relativePath, token, reason)
			}
		}
	}
}

func TestConfigValidationReportDocumentsTaskSixTwoEvidence(t *testing.T) {
	root := findRepoRoot(t)
	content := readRepoFile(t, root, ".kiro/specs/phase-01-config-system/validation.md")

	required := []string{
		"任务 6.2",
		"go test -count=1 ./...",
		"npm --prefix frontend run test",
		"npm --prefix frontend run typecheck",
		"范围扫描",
		"未实现完整设置页",
		"未创建 " + "SQL" + "ite " + "sche" + "ma",
		"未访问目标数据库",
		"未进行网络上传",
	}

	for _, text := range required {
		if !strings.Contains(content, text) {
			t.Fatalf("validation.md must mention %q", text)
		}
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repository root from %s", dir)
		}
		dir = parent
	}
}

func readRepoFile(t *testing.T, root string, relativePath string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err != nil {
		t.Fatalf("read %s: %v", relativePath, err)
	}
	return string(content)
}
