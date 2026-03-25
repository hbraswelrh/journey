// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"github.com/hbraswelrh/gemara-user-journey/internal/mcp"
	"github.com/hbraswelrh/gemara-user-journey/internal/schema"
	"github.com/hbraswelrh/gemara-user-journey/internal/session"
	"github.com/hbraswelrh/gemara-user-journey/internal/tutorials"
)

// UserPrompter abstracts user input for testing.
type UserPrompter interface {
	// Ask presents a question with options and returns the
	// selected option index.
	Ask(question string, options []string) (int, error)
	// AskText presents a question and returns free-text
	// input from the user.
	AskText(question string) (string, error)
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

	// Auto-select the latest schema version without user
	// interaction. The interactive RunVersionSelection
	// prompt is intentionally bypassed here. See
	// ADR-0003 for rationale. To re-enable version
	// selection, replace this block with a call to
	// RunVersionSelection.
	if cfg.VersionFetcher != nil {
		selRes, err := schema.AutoSelectLatest(
			ctx,
			cfg.VersionFetcher,
			cfg.VersionCachePath,
			result.Session,
		)
		if err != nil {
			// Non-fatal: proceed without a version
			// constraint and warn the user.
			fmt.Fprintln(out, RenderWarning(
				"Schema version could not be "+
					"resolved; proceeding without "+
					"version constraint.",
			))
		} else {
			fmt.Fprintln(out, RenderSuccess(
				fmt.Sprintf(
					"Schema version: %s (latest)",
					selRes.SelectedTag,
				),
			))
			if len(selRes.ExperimentalSchemas) > 0 {
				fmt.Fprintln(out, RenderNote(
					fmt.Sprintf(
						"The following schemas are "+
							"experimental: %s",
						strings.Join(
							selRes.ExperimentalSchemas,
							", ",
						),
					),
				))
			}
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
				len(roleResult.Profile.Recommendations),
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

		// Determine mode from existing config.
		mode := consts.MCPModeDefault
		binaryPath := detection.BinaryPath
		existingConfig, readErr := mcp.ReadOpenCodeConfig(
			cfg.ConfigPath,
		)
		if readErr == nil {
			if entry, ok := existingConfig.MCP[consts.MCPServerName]; ok {
				mode = mcp.ParseMCPMode(entry)
				if binaryPath == "" {
					binaryPath = mcp.MCPBinaryPath(
						entry,
					)
				}
			}
		}

		// Show how to update/rebuild from source.
		if binaryPath != "" {
			renderUpdateGuidance(out, binaryPath)
		}

		sess := session.NewSessionWithMCP("", mode)
		return &SetupResult{
			Session: sess,
		}, nil
	}

	// Step 2: Offer setup.
	fmt.Fprintln(out, subtleStyle.Render(
		"The Gemara MCP server must be installed "+
			"from source. If you have already "+
			"built it, provide the path to the "+
			"binary.",
	))
	fmt.Fprintln(out)

	choice, err := cfg.Prompter.Ask(
		"How would you like to set up the Gemara "+
			"MCP server?",
		[]string{
			"Build from source (clone and build)",
			"I already have it — provide the path",
			"Skip for now",
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
		// User has an existing build.
		return handleExistingBinary(cfg, out)
	default:
		// Declined.
		return handleDecline(out)
	}
}

func handleExistingBinary(
	cfg *SetupConfig,
	out io.Writer,
) (*SetupResult, error) {
	binaryPath, err := cfg.Prompter.AskText(
		"Path to gemara-mcp binary " +
			"(e.g., /path/to/gemara-mcp/bin/gemara-mcp):",
	)
	if err != nil {
		return nil, fmt.Errorf("prompt path: %w", err)
	}
	binaryPath = strings.TrimSpace(binaryPath)
	if binaryPath == "" {
		return handleDecline(out)
	}

	fmt.Fprintln(out, RenderSuccess(
		"Using existing binary: "+binaryPath,
	))

	// Prompt for server mode.
	mode, err := promptServerMode(cfg.Prompter, out)
	if err != nil {
		return nil, err
	}

	// Show config preview and write.
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"MCP Configuration",
	))
	fmt.Fprintln(out)
	configPreview := fmt.Sprintf(
		"command: %s\nargs:    [serve, --mode, %s]",
		binaryPath, mode,
	)
	fmt.Fprintln(out, codeBlockStyle.Render(
		configPreview,
	))
	fmt.Fprintln(out)

	if err := configureMCPEntry(
		cfg.ConfigPath, binaryPath, mode,
	); err != nil {
		return nil, err
	}
	fmt.Fprintln(out, RenderSuccess(
		"MCP configuration updated",
	))

	// Show update guidance.
	renderUpdateGuidance(out, binaryPath)

	sess := session.NewSessionWithMCP("", mode)
	return &SetupResult{
		Session:      sess,
		MCPInstalled: true,
	}, nil
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
	destDir := homeDir + "/.local/share/gemara-user-journey"

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

	// Prompt for server mode.
	mode, err := promptServerMode(cfg.Prompter, out)
	if err != nil {
		return nil, err
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
	configPreview := fmt.Sprintf(
		"command: %s\nargs:    [serve, --mode, %s]",
		binaryPath, mode,
	)
	fmt.Fprintln(out)
	fmt.Fprintln(out, codeBlockStyle.Render(
		configPreview,
	))
	fmt.Fprintln(out)

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
			cfg.ConfigPath, binaryPath, mode,
		); err != nil {
			return nil, err
		}
		fmt.Fprintln(out, RenderSuccess(
			"MCP configuration updated",
		))
	}

	sess := session.NewSessionWithMCP("", mode)
	return &SetupResult{
		Session:      sess,
		MCPInstalled: true,
	}, nil
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

// promptServerMode asks the user to select the MCP server
// operating mode (advisory or artifact).
func promptServerMode(
	prompter UserPrompter,
	out io.Writer,
) (string, error) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"Server Mode",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, subtleStyle.Render(
		"The Gemara MCP server supports two "+
			"operating modes:",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, subtleStyle.Render(
		"  Artifact — Advisory plus guided "+
			"creation wizards (tools + resources "+
			"+ prompts)",
	))
	fmt.Fprintln(out, subtleStyle.Render(
		"  Advisory — Read-only analysis and "+
			"validation (tools + resources only)",
	))
	fmt.Fprintln(out)

	choice, err := prompter.Ask(
		"Select server mode:",
		[]string{
			"Artifact (recommended, full capabilities)",
			"Advisory (read-only, no wizards)",
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"prompt server mode: %w", err,
		)
	}

	if choice == 1 {
		fmt.Fprintln(out, RenderNote(
			"Advisory mode selected. The "+
				"threat_assessment and "+
				"control_catalog prompts will not "+
				"be available. You can change the "+
				"mode later in opencode.json.",
		))
		return consts.MCPModeAdvisory, nil
	}
	return consts.MCPModeArtifact, nil
}

// renderUpdateGuidance shows the user how to keep their
// gemara-mcp build up to date from upstream source.
func renderUpdateGuidance(out io.Writer, binaryPath string) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"Keeping gemara-mcp Up to Date",
	))
	fmt.Fprintln(out)
	fmt.Fprintln(out, subtleStyle.Render(
		"To rebuild after upstream changes:"),
	)
	fmt.Fprintln(out)

	// Infer the repo directory from the binary path
	// (typically .../gemara-mcp/bin/gemara-mcp).
	repoDir := inferRepoDir(binaryPath)

	updateCmds := fmt.Sprintf(
		"cd %s\n"+
			"git fetch origin\n"+
			"git checkout main\n"+
			"git pull origin main\n"+
			"make build",
		repoDir,
	)
	fmt.Fprintln(out, codeBlockStyle.Render(updateCmds))
	fmt.Fprintln(out)
	fmt.Fprintln(out, faintStyle.Render(
		"Run ./gemara-user-journey --doctor to verify your "+
			"environment after rebuilding.",
	))
}

// inferRepoDir attempts to determine the gemara-mcp
// repository directory from the binary path. If the path
// ends in /bin/gemara-mcp, returns the parent of /bin.
// Otherwise returns the directory containing the binary.
func inferRepoDir(binaryPath string) string {
	// Common case: /path/to/gemara-mcp/bin/gemara-mcp
	if strings.HasSuffix(binaryPath, "/bin/gemara-mcp") {
		return strings.TrimSuffix(
			binaryPath, "/bin/gemara-mcp",
		)
	}
	// Fallback: directory containing the binary.
	idx := strings.LastIndex(binaryPath, "/")
	if idx > 0 {
		return binaryPath[:idx]
	}
	return "."
}

func configureMCPEntry(
	configPath string,
	binaryPath string,
	mode string,
) error {
	config, err := mcp.ReadOpenCodeConfig(configPath)
	if err != nil {
		return fmt.Errorf("read opencode config: %w", err)
	}
	mcp.EnsureMCPEntry(config, binaryPath, mode)
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
