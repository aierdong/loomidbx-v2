package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	// defaultAppName is used when callers do not provide an application directory name.
	defaultAppName = "LoomiDBX"

	// defaultConfigFileName is the ordinary JSON config file name inside the resolved config directory.
	defaultConfigFileName = "config.json"

	// defaultDataDirName is the leaf directory used for local app data below the resolved data root.
	defaultDataDirName = "data"
)

// PathResolver resolves ordinary config and local data paths for the current runtime mode.
type PathResolver interface {
	// Resolve returns absolute config/data paths and field-level issues for unusable path inputs.
	Resolve(input PathResolveInput) (ResolvedPaths, []ConfigIssue)
}

// PathResolveInput carries runtime mode and optional deterministic roots for path resolution.
type PathResolveInput struct {
	// AppName names the application directory placed below desktop roots.
	AppName string

	// Mode selects desktop, development, or test path semantics.
	Mode RunMode

	// ConfigDirOverride is an optional absolute directory used for config.json.
	ConfigDirOverride string

	// DataDirOverride is an optional absolute directory used for local data.
	DataDirOverride string

	// TestRoot is an optional absolute root for test-isolated config and data directories.
	TestRoot string

	// DevelopmentRoot is an optional absolute root for development-isolated config and data directories.
	DevelopmentRoot string

	// DesktopConfigRoot overrides the OS config root for deterministic tests or controlled desktop runs.
	DesktopConfigRoot string

	// DesktopDataRoot overrides the OS data root for deterministic tests or controlled desktop runs.
	DesktopDataRoot string
}

// ResolvedPaths contains the absolute config file and data paths accepted by the resolver.
type ResolvedPaths struct {
	// ConfigDir is the directory containing the ordinary app config file.
	ConfigDir string

	// ConfigFile is the ordinary app config file path.
	ConfigFile string

	// DataDir is the local application data directory for downstream storage specs.
	DataDir string

	// Mode is the runtime mode used to choose path semantics.
	Mode RunMode

	// Isolated reports whether the paths are intentionally separated from desktop defaults.
	Isolated bool
}

// DefaultPathResolver implements OS-aware path resolution using only the Go standard library.
type DefaultPathResolver struct{}

// Resolve returns deterministic absolute config and data paths for desktop, development, and test modes.
//
// The input may inject roots for tests or development tooling. Directory candidates are validated by
// creating the directory when necessary and writing a small probe file inside it. Failures are returned as
// ConfigIssue values with sanitized messages, so callers can reject invalid overrides without exposing full paths.
func (DefaultPathResolver) Resolve(input PathResolveInput) (ResolvedPaths, []ConfigIssue) {
	normalized := normalizePathInput(input)
	candidates := pathCandidates(normalized)

	configDir, configIssues := resolveWritableDir("paths.configDir", candidates.configDir)
	dataDir, dataIssues := resolveWritableDir("paths.dataDir", candidates.dataDir)

	issues := append(configIssues, dataIssues...)
	if len(issues) != 0 {
		return ResolvedPaths{
			ConfigDir:  configDir,
			ConfigFile: joinConfigFile(configDir),
			DataDir:    dataDir,
			Mode:       normalized.Mode,
			Isolated:   candidates.isolated,
		}, issues
	}

	return ResolvedPaths{
		ConfigDir:  configDir,
		ConfigFile: joinConfigFile(configDir),
		DataDir:    dataDir,
		Mode:       normalized.Mode,
		Isolated:   candidates.isolated,
	}, nil
}

// pathCandidateSet stores candidate directories before they are validated for absolute, creatable, and writable semantics.
type pathCandidateSet struct {
	// configDir is the candidate directory that should contain config.json.
	configDir string

	// dataDir is the candidate local data directory.
	dataDir string

	// isolated reports whether the candidates are intentionally outside normal desktop defaults.
	isolated bool
}

func normalizePathInput(input PathResolveInput) PathResolveInput {
	if input.AppName == "" {
		input.AppName = defaultAppName
	}
	if input.Mode == "" {
		input.Mode = ModeDesktop
	}
	return input
}

func pathCandidates(input PathResolveInput) pathCandidateSet {
	candidates := desktopPathCandidates(input)

	switch input.Mode {
	case ModeDevelopment:
		if input.DevelopmentRoot != "" {
			candidates.configDir = filepath.Join(input.DevelopmentRoot, "config")
			candidates.dataDir = filepath.Join(input.DevelopmentRoot, defaultDataDirName)
			candidates.isolated = true
		}
	case ModeTest:
		testRoot := input.TestRoot
		if testRoot == "" {
			testRoot = filepath.Join(os.TempDir(), "loomidbx-test")
		}
		candidates.configDir = filepath.Join(testRoot, "config")
		candidates.dataDir = filepath.Join(testRoot, defaultDataDirName)
		candidates.isolated = true
		return candidates
	}

	if input.ConfigDirOverride != "" {
		candidates.configDir = input.ConfigDirOverride
		candidates.isolated = true
	}
	if input.DataDirOverride != "" {
		candidates.dataDir = input.DataDirOverride
		candidates.isolated = true
	}

	return candidates
}

func desktopPathCandidates(input PathResolveInput) pathCandidateSet {
	configRoot := input.DesktopConfigRoot
	if configRoot == "" {
		configRoot = userConfigRoot()
	}

	dataRoot := input.DesktopDataRoot
	if dataRoot == "" {
		dataRoot = userDataRoot()
	}

	return pathCandidateSet{
		configDir: filepath.Join(configRoot, input.AppName),
		dataDir:   filepath.Join(dataRoot, input.AppName, defaultDataDirName),
	}
}

func userConfigRoot() string {
	if root, err := os.UserConfigDir(); err == nil && root != "" {
		return root
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".config")
	}
	return filepath.Join(os.TempDir(), "loomidbx-config")
}

func userDataRoot() string {
	switch runtime.GOOS {
	case "windows":
		if root := os.Getenv("LOCALAPPDATA"); root != "" {
			return root
		}
		if root := os.Getenv("APPDATA"); root != "" {
			return root
		}
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, "Library", "Application Support")
		}
	default:
		if root := os.Getenv("XDG_DATA_HOME"); root != "" {
			return root
		}
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, ".local", "share")
		}
	}
	if root, err := os.UserConfigDir(); err == nil && root != "" {
		return root
	}
	return filepath.Join(os.TempDir(), "loomidbx-data")
}

func resolveWritableDir(fieldPath string, candidate string) (string, []ConfigIssue) {
	if candidate == "" {
		return "", []ConfigIssue{invalidPathIssue(fieldPath, "路径不能为空")}
	}
	if !filepath.IsAbs(candidate) {
		return candidate, []ConfigIssue{invalidPathIssue(fieldPath, "路径必须是绝对路径")}
	}

	cleaned := filepath.Clean(candidate)
	if err := os.MkdirAll(cleaned, 0o700); err != nil {
		return cleaned, []ConfigIssue{invalidPathIssue(fieldPath, "路径不可创建或不可访问")}
	}

	info, err := os.Stat(cleaned)
	if err != nil {
		return cleaned, []ConfigIssue{invalidPathIssue(fieldPath, "路径不可访问")}
	}
	if !info.IsDir() {
		return cleaned, []ConfigIssue{invalidPathIssue(fieldPath, "路径必须指向目录")}
	}

	probe, err := os.CreateTemp(cleaned, ".loomidbx-write-check-*")
	if err != nil {
		return cleaned, []ConfigIssue{invalidPathIssue(fieldPath, "路径不可写")}
	}
	probeName := probe.Name()
	if closeErr := probe.Close(); closeErr != nil {
		_ = os.Remove(probeName)
		return cleaned, []ConfigIssue{invalidPathIssue(fieldPath, "路径不可写")}
	}
	if removeErr := os.Remove(probeName); removeErr != nil {
		return cleaned, []ConfigIssue{invalidPathIssue(fieldPath, "路径不可写")}
	}

	return cleaned, nil
}

func joinConfigFile(configDir string) string {
	if configDir == "" {
		return ""
	}
	return filepath.Join(configDir, defaultConfigFileName)
}

func invalidPathIssue(fieldPath string, message string) ConfigIssue {
	return ConfigIssue{
		Path:     fieldPath,
		Code:     ConfigIssueCodeConfigPathInvalid,
		Severity: ConfigIssueSeverityError,
		Message:  message,
	}
}
