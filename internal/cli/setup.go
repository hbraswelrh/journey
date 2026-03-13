// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/mcp"
	"github.com/hbraswelrh/pacman/internal/session"
	"github.com/hbraswelrh/pacman/internal/tutorials"
)

// UserPrompter abstracts user input for testing.
type UserPrompter interface {
	// Ask presents a question with options and returns the
	// selected option index.
	Ask(question string, options []string) (int, error)
}

// SetupConfig holds the dependencies for the setup flow.
type SetupConfig struct {
	// Prompter handles user interaction.
	Prompter UserPrompter
	// BinaryLookup checks if a binary is in PATH.
	BinaryLookup mcp.BinaryLookup
	// PodmanChecker checks for running Podman containers.
	PodmanChecker mcp.PodmanChecker
	// Installer handles gemara-mcp installation.
	Installer *mcp.Installer
	// SSHChecker detects whether SSH keys are configured
	// for GitHub. When nil, HTTPS is used by default.
	SSHChecker mcp.SSHChecker
	// ConfigPath is the path to opencode.json.
	ConfigPath string
	// VersionFetcher fetches releases from upstream. When
	// set, RunSetup will run version selection after MCP
	// setup completes.
	VersionFetcher ReleaseFetcherFn
	// VersionCachePath is the path to the local release
	// cache file.
	VersionCachePath string
	// RolePrompter handles free-text input for role
	// discovery. When set (along with TutorialsDir),
	// RunSetup will run role discovery after version
	// selection.
	RolePrompter FreeTextPrompter
	// TutorialsDir is the path to the Gemara tutorials
	// directory. Required for role discovery.
	TutorialsDir string
}

// SetupResult holds the outcome of the setup flow.
type SetupResult struct {
	// Session is the configured session.
	Session *session.Session
	// MCPInstalled is true if the MCP server was installed
	// during setup.
	MCPInstalled bool
	// Declined is true if the user declined MCP installation.
	Declined bool
}

// RunSetup executes the first-launch setup flow:
//  1. Detect whether gemara-mcp is already installed.
//  2. If not detected, explain the MCP tools and offer
//     installation (source build or Podman) or decline.
//  3. If installed or after installation, configure the
//     session with MCP connected.
//  4. If declined, configure the session in fallback mode
//     and list degraded capabilities.
//  5. If VersionFetcher is set, run schema version
//     selection.
func RunSetup(
	ctx context.Context,
	cfg *SetupConfig,
	out io.Writer,
) (*SetupResult, error) {
	result, err := runMCPSetup(ctx, cfg, out)
	if err != nil {
		return nil, err
	}

	// Run version selection if configured.
	if cfg.VersionFetcher != nil {
		vCfg := &VersionPromptConfig{
			Prompter:  cfg.Prompter,
			Fetcher:   cfg.VersionFetcher,
			CachePath: cfg.VersionCachePath,
			Session:   result.Session,
		}
		if err := RunVersionSelection(
			ctx, vCfg, out,
		); err != nil {
			return nil, fmt.Errorf(
				"version selection: %w", err,
			)
		}
	}

	// Run role discovery if configured.
	if cfg.RolePrompter != nil {
		roleCfg := &RolePromptConfig{
			Prompter:      cfg.RolePrompter,
			TutorialsDir:  cfg.TutorialsDir,
			SchemaVersion: result.Session.SchemaVersion,
		}
		roleResult, err := RunRoleDiscovery(
			roleCfg, out,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"role discovery: %w", err,
			)
		}

		// Store role profile in session.
		if roleResult.Profile != nil {
			roleName := ""
			if roleResult.Profile.Role != nil {
				roleName = roleResult.Profile.
					Role.Name
			}
			pathSteps := 0
			if roleResult.Tutorials != nil {
				path := generateLearningPath(
					roleResult, out,
				)
				if path != nil {
					pathSteps = len(path.Steps)
				}
			}
			result.Session.SetRoleProfile(
				roleName,
				roleResult.Profile.ExtractedKeywords,
				roleResult.Profile.
					UniqueLayerNumbers(),
				pathSteps,
			)
		}
	}

	return result, nil
}

// generateLearningPath builds and displays the learning path.
func generateLearningPath(
	roleResult *RolePromptResult,
	out io.Writer,
) *tutorials.LearningPath {
	if roleResult.Profile == nil ||
		len(roleResult.Tutorials) == 0 {
		return nil
	}

	path := tutorials.GeneratePath(
		roleResult.Profile,
		roleResult.Tutorials,
		"", // Version already checked in RolePromptResult
	)

	if len(path.Steps) > 0 {
		RenderLearningPath(path, out)
	}

	// Report missing layers.
	covered := make(map[int]bool)
	for _, step := range path.Steps {
		covered[step.Layer] = true
	}
	for _, layer := range roleResult.Profile.
		UniqueLayerNumbers() {
		if !covered[layer] {
			fmt.Fprintln(out, RenderNote(
				tutorials.MissingLayerMessage(layer),
			))
		}
	}

	return path
}

// runMCPSetup handles MCP detection and installation.
func runMCPSetup(
	ctx context.Context,
	cfg *SetupConfig,
	out io.Writer,
) (*SetupResult, error) {
	// Step 1: Detect existing installation.
	detection, err := mcp.Detect(
		cfg.BinaryLookup,
		cfg.PodmanChecker,
	)
	if err != nil {
		return nil, fmt.Errorf("detect MCP server: %w", err)
	}

	if detection.Detected {
		fmt.Fprintln(out, RenderSuccess(
			fmt.Sprintf(
				"Gemara MCP server detected (%s)",
				methodLabel(detection.Method),
			),
		))

		sess := session.NewSessionWithMCP("")
		return &SetupResult{
			Session: sess,
		}, nil
	}

	// Step 2: Explain and offer installation.
	fmt.Fprintln(out, RenderMCPToolsPanel())

	choice, err := cfg.Prompter.Ask(
		"How would you like to install the Gemara MCP server?",
		[]string{
			"Build from source (recommended)",
			"Run via Podman",
			"Skip installation",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("prompt user: %w", err)
	}

	switch choice {
	case 0:
		// Build from source.
		return handleSourceBuild(ctx, cfg, out)
	case 1:
		// Podman.
		return handlePodmanInstall(ctx, cfg, out)
	default:
		// Declined.
		return handleDecline(out)
	}
}

func handleSourceBuild(
	ctx context.Context,
	cfg *SetupConfig,
	out io.Writer,
) (*SetupResult, error) {
	// Auto-detect clone method: use SSH if keys are
	// configured, otherwise default to HTTPS.
	method := mcp.CloneHTTPS
	if cfg.SSHChecker != nil {
		fmt.Fprintln(out, RenderStatus(
			"Detecting SSH key configuration...",
		))
		method = mcp.DetectCloneMethod(
			ctx, cfg.SSHChecker,
		)
	}
	fmt.Fprintln(out, RenderSuccess(fmt.Sprintf(
		"Using %s for repository access",
		mcp.CloneMethodLabel(method),
	)))

	fmt.Fprintln(out, RenderStatus(
		"Resolving latest gemara-mcp release...",
	))

	release, err := cfg.Installer.ResolveLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve release: %w", err)
	}

	releaseLabel := fmt.Sprintf(
		"Found release %s (commit %s)",
		release.Tag,
		truncateSHA(release.CommitSHA),
	)
	if release.Prerelease {
		releaseLabel += " [prerelease]"
	}
	fmt.Fprintln(out, RenderSuccess(releaseLabel))
	fmt.Fprintln(out, RenderStatus(
		"Cloning and building...",
	))

	homeDir, err := userHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	destDir := homeDir + "/.local/share/pacman"

	binaryPath, err := cfg.Installer.CloneAndBuild(
		ctx, method, release, destDir,
	)
	if err != nil {
		return nil, fmt.Errorf("install: %w", err)
	}

	fmt.Fprintln(out, RenderSuccess(
		"Build complete: "+binaryPath,
	))

	// Save installed release metadata for future
	// update checks.
	installed := &mcp.InstalledRelease{
		Tag:        release.Tag,
		CommitSHA:  release.CommitSHA,
		Prerelease: release.Prerelease,
		InstalledAt: time.Now().UTC().Format(
			time.RFC3339,
		),
		BinaryPath: binaryPath,
	}
	if err := mcp.SaveInstalledRelease(
		destDir, installed,
	); err != nil {
		fmt.Fprintln(out, RenderWarning(
			"Could not save release metadata: "+
				err.Error(),
		))
	}

	// Show the user the absolute path and ask for
	// confirmation before writing MCP config.
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"MCP Configuration",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, subtleStyle.Render(
		"The following MCP server config will be "+
			"written:",
	))
	fmt.Fprintf(out, "\n  command: %s\n", binaryPath)
	fmt.Fprintf(out, "  args:    [serve]\n\n")

	confirmChoice, err := cfg.Prompter.Ask(
		"Write this MCP configuration?",
		[]string{
			"Yes, configure MCP server",
			"Skip configuration",
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"prompt config confirm: %w", err,
		)
	}
	if confirmChoice != 0 {
		fmt.Fprintln(out, RenderNote(
			"MCP configuration skipped. You can "+
				"configure it later by adding the "+
				"entry to your MCP client config.",
		))
	} else {
		if err := configureMCPEntry(
			cfg.ConfigPath, binaryPath,
		); err != nil {
			return nil, err
		}
		fmt.Fprintln(out, RenderSuccess(
			"MCP configuration updated",
		))
	}

	sess := session.NewSessionWithMCP("")
	return &SetupResult{
		Session:      sess,
		MCPInstalled: true,
	}, nil
}

func handlePodmanInstall(
	ctx context.Context,
	cfg *SetupConfig,
	out io.Writer,
) (*SetupResult, error) {
	fmt.Fprintln(out, RenderStatus(
		"Starting container...",
	))

	if err := cfg.Installer.InstallPodman(ctx); err != nil {
		return nil, fmt.Errorf("container install: %w", err)
	}

	fmt.Fprintln(out, RenderSuccess(
		"Container running",
	))

	// Detect container runtime (podman or docker).
	runtime := detectContainerRuntime(cfg)

	// Configure MCP entry for container-based server.
	config, err := mcp.ReadOpenCodeConfig(cfg.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf(
			"read config: %w", err,
		)
	}
	mcp.EnsureMCPEntryPodman(config, runtime)

	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"MCP Configuration",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, subtleStyle.Render(
		"The following MCP server config will be "+
			"written:",
	))
	fmt.Fprintf(out, "\n  command: %s\n", runtime)
	fmt.Fprintf(out,
		"  args:    [run, --rm, -i, %s, serve]\n\n",
		consts.MCPPodmanImage,
	)

	if err := mcp.WriteOpenCodeConfig(
		cfg.ConfigPath, config,
	); err != nil {
		return nil, fmt.Errorf(
			"write config: %w", err,
		)
	}
	fmt.Fprintln(out, RenderSuccess(
		"MCP configuration updated",
	))

	sess := session.NewSessionWithMCP("")
	return &SetupResult{
		Session:      sess,
		MCPInstalled: true,
	}, nil
}

// detectContainerRuntime returns "podman" if podman is
// available, otherwise "docker".
func detectContainerRuntime(cfg *SetupConfig) string {
	if _, err := cfg.BinaryLookup("podman"); err == nil {
		return "podman"
	}
	return "docker"
}

func handleDecline(out io.Writer) (*SetupResult, error) {
	sess := session.NewSessionWithoutMCP("")

	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderWarning(
		"MCP server installation skipped",
	))
	fmt.Fprintln(out, headingStyle.Render(
		"Degraded capabilities:",
	))
	for _, cap := range sess.DegradedCapabilities {
		fmt.Fprintf(out, "  %s %s\n",
			warningStyle.Render("▪"),
			cap,
		)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderNote(
		"You can install the MCP server at any "+
			"time during your session.",
	))

	return &SetupResult{
		Session:  sess,
		Declined: true,
	}, nil
}

func configureMCPEntry(
	configPath string,
	binaryPath string,
) error {
	config, err := mcp.ReadOpenCodeConfig(configPath)
	if err != nil {
		return fmt.Errorf("read opencode config: %w", err)
	}
	mcp.EnsureMCPEntry(config, binaryPath)
	if err := mcp.WriteOpenCodeConfig(
		configPath, config,
	); err != nil {
		return fmt.Errorf("write opencode config: %w", err)
	}
	return nil
}

func methodLabel(m mcp.InstallMethod) string {
	switch m {
	case mcp.MethodBinary:
		return "binary"
	case mcp.MethodPodman:
		return "podman"
	default:
		return "unknown"
	}
}

func truncateSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

// userHomeDir returns the user's home directory. Variable for
// testability.
var userHomeDir = os.UserHomeDir
