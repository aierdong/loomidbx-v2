package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type prerequisiteStatus string

const (
	// prerequisiteReady means the local tool exists and satisfies the configured minimum version.
	prerequisiteReady prerequisiteStatus = "ready"

	// prerequisiteMissing means the local tool command could not run.
	prerequisiteMissing prerequisiteStatus = "missing"

	// prerequisiteUnsupported means the local tool exists but does not satisfy the configured minimum version.
	prerequisiteUnsupported prerequisiteStatus = "unsupported"
)

type prerequisite struct {
	// Name is the developer-readable prerequisite label printed by doctor.
	Name string

	// Command is the executable name resolved from PATH.
	Command string

	// Args are passed to Command to retrieve version or diagnostic output.
	Args []string

	// Minimum is the optional minimum major.minor version required by this project.
	Minimum string

	// InstallHint explains the next local action when the prerequisite is missing or unsupported.
	InstallHint string

	// Required marks whether this prerequisite blocks the standard validation path.
	Required bool
}

type prerequisiteResult struct {
	// Status describes whether the prerequisite is ready, missing, or unsupported.
	Status prerequisiteStatus

	// Name is copied from the checked prerequisite for reporting.
	Name string

	// Version is the first line of successful command output.
	Version string

	// Diagnostic contains the command failure or version mismatch detail.
	Diagnostic string

	// Action contains the next local step the developer can take.
	Action string

	// Blocking is true when the result should make doctor fail.
	Blocking bool
}

func main() {
	checks := []prerequisite{
		{
			Name:        "Go 1.25+",
			Command:     "go",
			Args:        []string{"version"},
			Minimum:     "1.25",
			InstallHint: "安装 Go 1.25+：https://go.dev/doc/install",
			Required:    true,
		},
		{
			Name:        "Node.js",
			Command:     "node",
			Args:        []string{"--version"},
			InstallHint: "安装 Node.js LTS：https://nodejs.org/",
			Required:    true,
		},
		{
			Name:        "npm",
			Command:     "npm",
			Args:        []string{"--version"},
			InstallHint: "安装 npm：随 Node.js LTS 安装，或运行 corepack/npm 官方安装流程",
			Required:    true,
		},
		{
			Name:        "wails3",
			Command:     "wails3",
			Args:        []string{"doctor"},
			InstallHint: "安装 Wails v3 CLI：go install github.com/wailsapp/wails/v3/cmd/wails3@latest，并确认 GOPATH/bin 已加入 PATH；安装后可运行 wails3 doctor 查看平台依赖。",
			Required:    true,
		},
	}

	fmt.Println("LoomiDBX doctor")
	fmt.Println("检查本地开发工具链；不会读取或上传数据库连接、Schema、生成数据、Project 配置、用户 SQL 或远端账号数据。")
	fmt.Printf("平台：%s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	blockingFailure := false
	for _, check := range checks {
		output, err := run(check.Command, check.Args...)
		result := evaluatePrerequisite(check, output, err)
		printPrerequisiteResult(result)
		if result.Blocking {
			blockingFailure = true
		}
	}

	fmt.Println()
	printPlatformHint()

	if blockingFailure {
		fmt.Println("doctor 完成：存在缺失或版本不满足的前置工具。请按上述提示安装或升级后重试。")
		os.Exit(1)
	}

	fmt.Println("doctor 完成：基础工具链已就绪。")
}

func evaluatePrerequisite(check prerequisite, output string, err error) prerequisiteResult {
	result := prerequisiteResult{
		Name:   check.Name,
		Action: check.InstallHint,
	}
	if err != nil {
		result.Status = prerequisiteMissing
		result.Diagnostic = err.Error()
		result.Blocking = check.Required
		return result
	}

	result.Version = firstLine(output)
	if check.Minimum != "" && !meetsMinimumVersion(result.Version, check.Minimum) {
		result.Status = prerequisiteUnsupported
		result.Diagnostic = fmt.Sprintf("最低版本：%s；当前输出：%s", check.Minimum, result.Version)
		result.Blocking = check.Required
		return result
	}

	result.Status = prerequisiteReady
	return result
}

func printPrerequisiteResult(result prerequisiteResult) {
	switch result.Status {
	case prerequisiteMissing:
		fmt.Printf("[缺失] %s：%s\n", result.Name, result.Diagnostic)
		fmt.Printf("       下一步：%s\n", result.Action)
	case prerequisiteUnsupported:
		fmt.Printf("[版本不足] %s：%s\n", result.Name, result.Diagnostic)
		fmt.Printf("       下一步：%s\n", result.Action)
	default:
		fmt.Printf("[就绪] %s：%s\n", result.Name, result.Version)
	}
}

func run(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func firstLine(output string) string {
	if output == "" {
		return "命令执行成功"
	}
	line, _, _ := strings.Cut(output, "\n")
	return strings.TrimSpace(line)
}

func meetsMinimumVersion(version string, minimum string) bool {
	actualMajor, actualMinor, ok := extractMajorMinor(version)
	if !ok {
		return strings.Contains(version, minimum)
	}

	minimumMajor, minimumMinor, ok := parseMajorMinor(minimum)
	if !ok {
		return strings.Contains(version, minimum)
	}

	if actualMajor != minimumMajor {
		return actualMajor > minimumMajor
	}
	return actualMinor >= minimumMinor
}

func extractMajorMinor(version string) (int, int, bool) {
	for _, field := range strings.Fields(version) {
		field = strings.TrimPrefix(field, "go")
		if major, minor, ok := parseMajorMinor(field); ok {
			return major, minor, true
		}
	}
	return 0, 0, false
}

func parseMajorMinor(version string) (int, int, bool) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, false
	}

	major, ok := parseSmallInt(parts[0])
	if !ok {
		return 0, 0, false
	}
	minor, ok := parseSmallInt(parts[1])
	if !ok {
		return 0, 0, false
	}
	return major, minor, true
}

func parseSmallInt(value string) (int, bool) {
	result := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			break
		}
		result = result*10 + int(r-'0')
	}
	return result, value != "" && value[0] >= '0' && value[0] <= '9'
}

func printPlatformHint() {
	switch runtime.GOOS {
	case "windows":
		fmt.Println("平台提示：Windows 需要安装 WebView2 Runtime；如 Wails doctor 报缺失，请按其提示安装。")
	case "darwin":
		fmt.Println("平台提示：macOS 需要 Xcode Command Line Tools；如缺失请运行 xcode-select --install。")
	case "linux":
		fmt.Println("平台提示：Linux 需要 WebKitGTK 等桌面依赖；请按 Wails v3 文档安装发行版对应包。")
	default:
		fmt.Printf("平台提示：%s 可能需要额外桌面运行依赖，请参考 Wails v3 文档。\n", runtime.GOOS)
	}
}
