// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/mcp"
	"github.com/hbraswelrh/pacman/internal/session"
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
	// ConfigPath is the path to opencode.json.
	ConfigPath string
	// VersionFetcher fetches releases from upstream. When
	// set, RunSetup will run version selection after MCP
	// setup completes.
	VersionFetcher ReleaseFetcherFn
	// VersionCachePath is the path to the local release
	// cache file.
	VersionCachePath string
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

	return result, nil
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
		fmt.Fprintf(
			out,
			"Gemara MCP server detected (%s).\n",
			methodLabel(detection.Method),
		)

		// Configure opencode.json if binary was found.
		if detection.Method == mcp.MethodBinary {
			if err := configureMCPEntry(
				cfg.ConfigPath,
				detection.BinaryPath,
			); err != nil {
				return nil, err
			}
		}

		sess := session.NewSessionWithMCP("")
		return &SetupResult{
			Session: sess,
		}, nil
	}

	// Step 2: Explain and offer installation.
	fmt.Fprintf(out, "\n"+
		"The Gemara MCP server provides three tools that "+
		"enhance\nPac-Man's capabilities:\n\n"+
		"  - %s: Retrieve the upstream Gemara lexicon\n"+
		"    to ensure consistent terminology.\n"+
		"  - %s: Validate YAML artifacts against\n"+
		"    Gemara CUE schemas without local CUE tooling.\n"+
		"  - %s: Retrieve schema documentation for\n"+
		"    contextual reference during authoring.\n\n",
		consts.ToolGetLexicon,
		consts.ToolValidateArtifact,
		consts.ToolGetSchemaDocs,
	)

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
	// Ask for clone method.
	cloneChoice, err := cfg.Prompter.Ask(
		"Clone via SSH or HTTPS?",
		[]string{"SSH", "HTTPS"},
	)
	if err != nil {
		return nil, fmt.Errorf("prompt clone method: %w", err)
	}

	method := mcp.CloneHTTPS
	if cloneChoice == 0 {
		method = mcp.CloneSSH
	}

	fmt.Fprintf(out, "Resolving latest gemara-mcp release...\n")

	release, err := cfg.Installer.ResolveLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve release: %w", err)
	}

	fmt.Fprintf(
		out,
		"Found release %s (commit %s).\n"+
			"Cloning and building...\n",
		release.Tag,
		truncateSHA(release.CommitSHA),
	)

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

	fmt.Fprintf(
		out,
		"Build complete: %s\n",
		binaryPath,
	)

	// Configure opencode.json.
	if err := configureMCPEntry(
		cfg.ConfigPath, binaryPath,
	); err != nil {
		return nil, err
	}

	fmt.Fprintf(
		out,
		"OpenCode MCP configuration updated.\n",
	)

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
	fmt.Fprintf(out, "Starting Podman container...\n")

	if err := cfg.Installer.InstallPodman(ctx); err != nil {
		return nil, fmt.Errorf("podman install: %w", err)
	}

	fmt.Fprintf(out, "Podman container running.\n")

	sess := session.NewSessionWithMCP("")
	return &SetupResult{
		Session:      sess,
		MCPInstalled: true,
	}, nil
}

func handleDecline(out io.Writer) (*SetupResult, error) {
	sess := session.NewSessionWithoutMCP("")

	fmt.Fprintf(out, "\nMCP server installation skipped.\n")
	fmt.Fprintf(out, "Degraded capabilities:\n")
	for _, cap := range sess.DegradedCapabilities {
		fmt.Fprintf(out, "  - %s\n", cap)
	}
	fmt.Fprintf(
		out,
		"\nYou can install the MCP server at any time "+
			"during your session.\n",
	)

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
