package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPathResolverUsesDesktopDefaultRoots(t *testing.T) {
	root := t.TempDir()
	configRoot := filepath.Join(root, "config-root")
	dataRoot := filepath.Join(root, "data-root")

	resolved, issues := DefaultPathResolver{}.Resolve(PathResolveInput{
		AppName:           "LoomiDBX",
		Mode:              ModeDesktop,
		DesktopConfigRoot: configRoot,
		DesktopDataRoot:   dataRoot,
	})

	assertNoPathIssues(t, issues)
	assertAbsPath(t, resolved.ConfigFile)
	assertAbsPath(t, resolved.ConfigDir)
	assertAbsPath(t, resolved.DataDir)
	assertPathEqual(t, resolved.ConfigDir, filepath.Join(configRoot, "LoomiDBX"))
	assertPathEqual(t, resolved.ConfigFile, filepath.Join(configRoot, "LoomiDBX", "config.json"))
	assertPathEqual(t, resolved.DataDir, filepath.Join(dataRoot, "LoomiDBX", "data"))
	if resolved.Mode != ModeDesktop {
		t.Fatalf("Mode = %q, want %q", resolved.Mode, ModeDesktop)
	}
}

func TestPathResolverUsesDevelopmentDirectoryOverrides(t *testing.T) {
	root := t.TempDir()
	configOverride := filepath.Join(root, "dev-config")
	dataOverride := filepath.Join(root, "dev-data")

	resolved, issues := DefaultPathResolver{}.Resolve(PathResolveInput{
		AppName:           "LoomiDBX",
		Mode:              ModeDevelopment,
		ConfigDirOverride: configOverride,
		DataDirOverride:   dataOverride,
	})

	assertNoPathIssues(t, issues)
	assertPathEqual(t, resolved.ConfigDir, configOverride)
	assertPathEqual(t, resolved.ConfigFile, filepath.Join(configOverride, "config.json"))
	assertPathEqual(t, resolved.DataDir, dataOverride)
	if !resolved.Isolated {
		t.Fatal("development override paths should be isolated from desktop defaults")
	}
}

func TestPathResolverUsesTestIsolationRoot(t *testing.T) {
	testRoot := t.TempDir()

	resolved, issues := DefaultPathResolver{}.Resolve(PathResolveInput{
		AppName:  "LoomiDBX",
		Mode:     ModeTest,
		TestRoot: testRoot,
	})

	assertNoPathIssues(t, issues)
	assertPathEqual(t, resolved.ConfigDir, filepath.Join(testRoot, "config"))
	assertPathEqual(t, resolved.ConfigFile, filepath.Join(testRoot, "config", "config.json"))
	assertPathEqual(t, resolved.DataDir, filepath.Join(testRoot, "data"))
	if !resolved.Isolated {
		t.Fatal("test paths must be isolated")
	}
}

func TestPathResolverKeepsTestModeIsolatedWhenOverridesArePresent(t *testing.T) {
	root := t.TempDir()
	testRoot := filepath.Join(root, "test-root")
	desktopConfigRoot := filepath.Join(root, "desktop-config")
	desktopDataRoot := filepath.Join(root, "desktop-data")

	resolved, issues := DefaultPathResolver{}.Resolve(PathResolveInput{
		AppName:           "LoomiDBX",
		Mode:              ModeTest,
		TestRoot:          testRoot,
		ConfigDirOverride: filepath.Join(desktopConfigRoot, "LoomiDBX"),
		DataDirOverride:   filepath.Join(desktopDataRoot, "LoomiDBX", "data"),
		DesktopConfigRoot: desktopConfigRoot,
		DesktopDataRoot:   desktopDataRoot,
	})

	assertNoPathIssues(t, issues)
	assertPathEqual(t, resolved.ConfigDir, filepath.Join(testRoot, "config"))
	assertPathEqual(t, resolved.ConfigFile, filepath.Join(testRoot, "config", "config.json"))
	assertPathEqual(t, resolved.DataDir, filepath.Join(testRoot, "data"))
	if !resolved.Isolated {
		t.Fatal("test paths must remain isolated even when overrides are present")
	}
}

func TestPathResolverReportsNonAbsoluteOverrideIssue(t *testing.T) {
	root := t.TempDir()

	_, issues := DefaultPathResolver{}.Resolve(PathResolveInput{
		AppName:           "LoomiDBX",
		Mode:              ModeDevelopment,
		ConfigDirOverride: "relative-config",
		DesktopConfigRoot: filepath.Join(root, "config-root"),
		DesktopDataRoot:   filepath.Join(root, "data-root"),
	})

	assertIssue(t, issues, "paths.configDir", ConfigIssueCodeConfigPathInvalid)
}

func TestPathResolverReportsUncreatablePathIssue(t *testing.T) {
	root := t.TempDir()
	fileAsDir := filepath.Join(root, "file-instead-of-dir")
	if err := os.WriteFile(fileAsDir, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, issues := DefaultPathResolver{}.Resolve(PathResolveInput{
		AppName:           "LoomiDBX",
		Mode:              ModeDevelopment,
		DataDirOverride:   filepath.Join(fileAsDir, "child"),
		DesktopConfigRoot: filepath.Join(root, "config-root"),
		DesktopDataRoot:   filepath.Join(root, "data-root"),
	})

	assertIssue(t, issues, "paths.dataDir", ConfigIssueCodeConfigPathInvalid)
	assertMessagesDoNotContain(t, issues, fileAsDir)
}

func TestPathResolverReportsUnwritablePathIssue(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows ACL semantics make chmod-based unwritable directory checks non-deterministic")
	}

	root := t.TempDir()
	readOnlyParent := filepath.Join(root, "readonly")
	if err := os.Mkdir(readOnlyParent, 0o500); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	defer func() {
		if err := os.Chmod(readOnlyParent, 0o700); err != nil {
			t.Fatalf("restore chmod error = %v", err)
		}
	}()

	_, issues := DefaultPathResolver{}.Resolve(PathResolveInput{
		AppName:           "LoomiDBX",
		Mode:              ModeDevelopment,
		ConfigDirOverride: filepath.Join(readOnlyParent, "config"),
		DesktopConfigRoot: filepath.Join(root, "config-root"),
		DesktopDataRoot:   filepath.Join(root, "data-root"),
	})

	assertIssue(t, issues, "paths.configDir", ConfigIssueCodeConfigPathInvalid)
	assertMessagesDoNotContain(t, issues, readOnlyParent)
}

func assertNoPathIssues(t *testing.T, issues []ConfigIssue) {
	t.Helper()
	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func assertAbsPath(t *testing.T, path string) {
	t.Helper()
	if !filepath.IsAbs(path) {
		t.Fatalf("path %q is not absolute", path)
	}
}

func assertPathEqual(t *testing.T, got string, want string) {
	t.Helper()
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func assertIssue(t *testing.T, issues []ConfigIssue, path string, code ConfigIssueCode) {
	t.Helper()
	for _, issue := range issues {
		if issue.Path == path && issue.Code == code && issue.Severity == ConfigIssueSeverityError {
			return
		}
	}
	t.Fatalf("issues = %+v, want %s issue at %s", issues, code, path)
}

func assertMessagesDoNotContain(t *testing.T, issues []ConfigIssue, forbidden string) {
	t.Helper()
	for _, issue := range issues {
		if strings.Contains(issue.Message, forbidden) {
			t.Fatalf("issue message %q contains sensitive path %q", issue.Message, forbidden)
		}
	}
}
