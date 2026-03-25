// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"github.com/hbraswelrh/gemara-user-journey/internal/mcp"
)

// CheckStatus represents the result of a single doctor check.
type CheckStatus int

const (
	// CheckPass means the check passed.
	CheckPass CheckStatus = iota
	// CheckWarn means the check passed with a warning.
	CheckWarn
	// CheckFail means the check failed.
	CheckFail
)

// CheckResult holds the outcome of a single doctor check.
type CheckResult struct {
	Name    string
	Status  CheckStatus
	Message string
	Fix     string
}

// DoctorConfig holds dependencies for the doctor command,
// allowing testability.
type DoctorConfig struct {
	// LookupBinary resolves binary names to paths.
	LookupBinary func(string) (string, error)
	// ReadConfig reads the OpenCode config file.
	ReadConfig func(string) (*mcp.OpenCodeConfig, error)
	// ConfigPath is the path to opencode.json.
	ConfigPath string
	// TutorialsDir is the resolved tutorials directory.
	TutorialsDir string
}

// DefaultDoctorConfig returns a DoctorConfig using real
// system lookups.
func DefaultDoctorConfig(
	configPath string,
) *DoctorConfig {
	homeDir, _ := os.UserHomeDir()
	tutDir := consts.DefaultTutorialsDir
	if homeDir != "" {
		tutDir = filepath.Join(
			homeDir,
			consts.DefaultGemaraDir,
			consts.GemaraTutorialsSubdir,
		)
	}
	return &DoctorConfig{
		LookupBinary: exec.LookPath,
		ReadConfig:   mcp.ReadOpenCodeConfig,
		ConfigPath:   configPath,
		TutorialsDir: tutDir,
	}
}

// RunDoctor performs all environment checks and writes
// a formatted report to out. Returns true if all critical
// checks pass.
func RunDoctor(
	cfg *DoctorConfig,
	out io.Writer,
) bool {
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"🩺 Gemara User Journey Doctor",
	))
	fmt.Fprintln(out, subtleStyle.Render(
		"Checking your environment...",
	))
	fmt.Fprintln(out)

	checks := runAllChecks(cfg)

	passed := 0
	warned := 0
	failed := 0

	for _, c := range checks {
		switch c.Status {
		case CheckPass:
			fmt.Fprintln(out, renderCheckPass(c))
			passed++
		case CheckWarn:
			fmt.Fprintln(out, renderCheckWarn(c))
			warned++
		case CheckFail:
			fmt.Fprintln(out, renderCheckFail(c))
			failed++
		}
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)

	summary := fmt.Sprintf(
		"%d passed, %d warnings, %d failed",
		passed, warned, failed,
	)
	if failed == 0 {
		fmt.Fprintln(out, successStyle.Render(
			"✅ All checks passed ("+summary+")",
		))

		// Show MCP capabilities reference.
		fmt.Fprintln(out)
		fmt.Fprintln(out, RenderMCPToolsPanel())

		// Prompt to start OpenCode.
		fmt.Fprintln(out, successStyle.Render(
			"🚀 Ready to go!",
		))
		fmt.Fprintln(out)
		fmt.Fprintln(out, subtleStyle.Render(
			"Start OpenCode from your project "+
				"directory to begin:",
		))
		fmt.Fprintln(out)
		fmt.Fprintln(out, codeBlockStyle.Render(
			"opencode",
		))
		fmt.Fprintln(out)
		fmt.Fprintln(out, faintStyle.Render(
			"OpenCode will start the gemara-mcp "+
				"server automatically and provide "+
				"access to all MCP capabilities "+
				"listed above.",
		))
	} else {
		fmt.Fprintln(out, dangerStyle.Render(
			"❌ Some checks failed ("+summary+")",
		))
		fmt.Fprintln(out)
		fmt.Fprintln(out, subtleStyle.Render(
			"Fix the issues above and run "+
				"./gemara-user-journey --doctor again.",
		))
	}
	fmt.Fprintln(out)

	return failed == 0
}

// runAllChecks executes every doctor check in order.
func runAllChecks(cfg *DoctorConfig) []CheckResult {
	return []CheckResult{
		checkOpenCode(cfg.LookupBinary),
		checkGo(cfg.LookupBinary),
		checkCUE(cfg.LookupBinary),
		checkGemaraMCP(cfg.LookupBinary),
		checkOpenCodeConfig(cfg.ReadConfig, cfg.ConfigPath),
		checkMCPMode(cfg.ReadConfig, cfg.ConfigPath),
		checkGit(cfg.LookupBinary),
		checkTutorials(cfg.TutorialsDir),
	}
}

// --- Individual Checks ---

func checkOpenCode(
	lookup func(string) (string, error),
) CheckResult {
	path, err := lookup("opencode")
	if err != nil || path == "" {
		return CheckResult{
			Name:    "💻 OpenCode",
			Status:  CheckFail,
			Message: "opencode not found in PATH",
			Fix: "Install OpenCode:\n" +
				"  macOS:  brew install " +
				"anomalyco/tap/opencode\n" +
				"  Linux:  curl -fsSL " +
				"https://opencode.ai/install | bash",
		}
	}
	return CheckResult{
		Name:    "💻 OpenCode",
		Status:  CheckPass,
		Message: path,
	}
}

func checkGo(
	lookup func(string) (string, error),
) CheckResult {
	path, err := lookup("go")
	if err != nil || path == "" {
		return CheckResult{
			Name:    "🔧 Go",
			Status:  CheckFail,
			Message: "go not found in PATH",
			Fix: "Install Go: " +
				"https://go.dev/dl/",
		}
	}
	return CheckResult{
		Name:    "🔧 Go",
		Status:  CheckPass,
		Message: path,
	}
}

func checkCUE(
	lookup func(string) (string, error),
) CheckResult {
	path, err := lookup("cue")
	if err != nil || path == "" {
		return CheckResult{
			Name:    "📐 CUE",
			Status:  CheckWarn,
			Message: "cue not found in PATH",
			Fix: "Install CUE:\n" +
				"  brew install cue-lang/tap/cue\n" +
				"  (required for local schema " +
				"validation when MCP is unavailable)",
		}
	}
	return CheckResult{
		Name:    "📐 CUE",
		Status:  CheckPass,
		Message: path,
	}
}

func checkGemaraMCP(
	lookup func(string) (string, error),
) CheckResult {
	path, err := lookup(consts.MCPBinaryName)
	if err != nil || path == "" {
		return CheckResult{
			Name:   "🔌 MCP Server",
			Status: CheckWarn,
			Message: consts.MCPBinaryName +
				" not found in PATH",
			Fix: "Install gemara-mcp from source:\n" +
				"  git clone " +
				consts.GemaraMCPCloneHTTPS + "\n" +
				"  cd gemara-mcp\n" +
				"  git checkout main\n" +
				"  make build\n" +
				"  (or install via Gemara User Journey's " +
				"setup flow)",
		}
	}
	return CheckResult{
		Name:    "🔌 MCP Server",
		Status:  CheckPass,
		Message: path,
	}
}

func checkOpenCodeConfig(
	readConfig func(string) (*mcp.OpenCodeConfig, error),
	configPath string,
) CheckResult {
	if configPath == "" {
		return CheckResult{
			Name:    "📄 opencode.json",
			Status:  CheckFail,
			Message: "no config path configured",
			Fix: "Create opencode.json in your " +
				"project root with a gemara-mcp " +
				"entry",
		}
	}

	config, err := readConfig(configPath)
	if err != nil {
		return CheckResult{
			Name:    "📄 opencode.json",
			Status:  CheckFail,
			Message: "cannot read: " + err.Error(),
			Fix: "Create opencode.json:\n" +
				"  {\n" +
				"    \"mcp\": {\n" +
				"      \"gemara-mcp\": {\n" +
				"        \"type\": \"local\",\n" +
				"        \"command\": " +
				"[\"/path/to/gemara-mcp\"," +
				" \"serve\", \"--mode\"," +
				" \"artifact\"],\n" +
				"        \"enabled\": true\n" +
				"      }\n" +
				"    }\n" +
				"  }",
		}
	}

	entry, ok := config.MCP[consts.MCPServerName]
	if !ok {
		return CheckResult{
			Name:   "📄 opencode.json",
			Status: CheckFail,
			Message: "no " + consts.MCPServerName +
				" entry found",
			Fix: "Add a gemara-mcp entry to " +
				"opencode.json with command and " +
				"args fields",
		}
	}

	if len(entry.Command) == 0 {
		return CheckResult{
			Name:    "📄 opencode.json",
			Status:  CheckFail,
			Message: "gemara-mcp entry has empty command",
			Fix: "Set the command array:\n" +
				"  [\"<path>/gemara-mcp\"," +
				" \"serve\", \"--mode\"," +
				" \"artifact\"]",
		}
	}

	// Verify the binary path exists on disk.
	binaryPath := entry.Command[0]
	if _, err := os.Stat(binaryPath); err != nil {
		return CheckResult{
			Name:   "📄 opencode.json",
			Status: CheckWarn,
			Message: "binary not found: " +
				binaryPath,
			Fix: "Build gemara-mcp from source:\n" +
				"  git clone " +
				consts.GemaraCloneHTTPS + "\n" +
				"  cd gemara-mcp && git " +
				"checkout main && make build\n" +
				"  Then update the path in " +
				"opencode.json",
		}
	}

	return CheckResult{
		Name:   "📄 opencode.json",
		Status: CheckPass,
		Message: consts.MCPServerName +
			" configured (" + binaryPath + ")",
	}
}

func checkMCPMode(
	readConfig func(string) (*mcp.OpenCodeConfig, error),
	configPath string,
) CheckResult {
	if configPath == "" {
		return CheckResult{
			Name:    "⚙️  Server Mode",
			Status:  CheckWarn,
			Message: "cannot check mode without config",
			Fix:     "Run the opencode.json check first",
		}
	}

	config, err := readConfig(configPath)
	if err != nil {
		return CheckResult{
			Name:    "⚙️  Server Mode",
			Status:  CheckWarn,
			Message: "cannot read config",
		}
	}

	entry, ok := config.MCP[consts.MCPServerName]
	if !ok {
		return CheckResult{
			Name:    "⚙️  Server Mode",
			Status:  CheckWarn,
			Message: "no gemara-mcp entry to check mode",
		}
	}

	mode := mcp.ParseMCPMode(entry)
	hasServe := false
	hasModeFlag := false
	for i, arg := range entry.Command {
		if arg == "serve" {
			hasServe = true
		}
		if arg == consts.MCPModeFlag &&
			i+1 < len(entry.Command) {
			hasModeFlag = true
		}
	}

	if !hasServe {
		return CheckResult{
			Name:    "⚙️  Server Mode",
			Status:  CheckWarn,
			Message: "args missing 'serve' subcommand",
			Fix: "Update args to: " +
				"[\"serve\", \"--mode\", \"artifact\"]",
		}
	}

	if !hasModeFlag {
		return CheckResult{
			Name:   "⚙️  Server Mode",
			Status: CheckWarn,
			Message: "no --mode flag; defaulting to " +
				consts.MCPModeDefault,
			Fix: "Add --mode to args: " +
				"[\"serve\", \"--mode\", \"artifact\"]",
		}
	}

	msg := mode
	if mode == consts.MCPModeArtifact {
		msg += " (wizard prompts available)"
	} else {
		msg += " (wizard prompts disabled)"
	}

	return CheckResult{
		Name:    "⚙️  Server Mode",
		Status:  CheckPass,
		Message: msg,
	}
}

func checkGit(
	lookup func(string) (string, error),
) CheckResult {
	path, err := lookup("git")
	if err != nil || path == "" {
		return CheckResult{
			Name:    "🔀 Git",
			Status:  CheckFail,
			Message: "git not found in PATH",
			Fix:     "Install Git: https://git-scm.com",
		}
	}
	return CheckResult{
		Name:    "🔀 Git",
		Status:  CheckPass,
		Message: path,
	}
}

func checkTutorials(tutorialsDir string) CheckResult {
	if tutorialsDir == "" {
		return CheckResult{
			Name:    "📚 Tutorials",
			Status:  CheckFail,
			Message: "no tutorials directory configured",
			Fix: "Clone the Gemara repository:\n" +
				"  git clone " +
				consts.GemaraCloneHTTPS + "\n" +
				"  Tutorials are at " +
				consts.GemaraTutorialsSubdir,
		}
	}

	info, err := os.Stat(tutorialsDir)
	if err != nil || !info.IsDir() {
		return CheckResult{
			Name:    "📚 Tutorials",
			Status:  CheckWarn,
			Message: "not found at " + tutorialsDir,
			Fix: "Gemara User Journey will clone the Gemara " +
				"repository automatically on " +
				"first tutorial launch, or clone " +
				"manually:\n" +
				"  git clone --branch main " +
				"--single-branch --depth 1 " +
				consts.GemaraCloneHTTPS,
		}
	}

	// Count tutorial files.
	entries, _ := os.ReadDir(tutorialsDir)
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			count++
		}
	}

	return CheckResult{
		Name:   "📚 Tutorials",
		Status: CheckPass,
		Message: fmt.Sprintf(
			"%s (%d categories)",
			tutorialsDir, count,
		),
	}
}

// --- Rendering ---

func renderCheckPass(c CheckResult) string {
	return "  ✅ " +
		fmt.Sprintf("%-22s", c.Name) +
		faintStyle.Render(c.Message)
}

func renderCheckWarn(c CheckResult) string {
	line := "  ⚠️  " +
		fmt.Sprintf("%-22s", c.Name) +
		warningStyle.Render(c.Message)
	if c.Fix != "" {
		for _, fixLine := range strings.Split(
			c.Fix, "\n",
		) {
			line += "\n" + faintStyle.Render(
				"                           "+
					fixLine,
			)
		}
	}
	return line
}

func renderCheckFail(c CheckResult) string {
	line := "  ❌ " +
		fmt.Sprintf("%-22s", c.Name) +
		dangerStyle.Render(c.Message)
	if c.Fix != "" {
		for _, fixLine := range strings.Split(
			c.Fix, "\n",
		) {
			line += "\n" + faintStyle.Render(
				"                           "+
					fixLine,
			)
		}
	}
	return line
}
