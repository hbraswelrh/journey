// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"charm.land/huh/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hbraswelrh/pacman/internal/cli"
	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/mcp"
	"github.com/hbraswelrh/pacman/internal/schema"
	"github.com/hbraswelrh/pacman/internal/session"
)

// huhPrompter implements cli.FreeTextPrompter using the
// charmbracelet/huh interactive widgets.
type huhPrompter struct{}

func (p *huhPrompter) Ask(
	question string,
	options []string,
) (int, error) {
	var selected int

	opts := make([]huh.Option[int], len(options))
	for i, label := range options {
		opts[i] = huh.NewOption(label, i)
	}

	err := huh.NewSelect[int]().
		Title(question).
		Options(opts...).
		Value(&selected).
		Run()
	if err != nil {
		return 0, fmt.Errorf("prompt: %w", err)
	}

	return selected, nil
}

func (p *huhPrompter) AskText(
	question string,
) (string, error) {
	var answer string

	err := huh.NewInput().
		Title(question).
		Value(&answer).
		Run()
	if err != nil {
		return "", fmt.Errorf("prompt: %w", err)
	}

	return answer, nil
}

// demoPrompter simulates user choices for non-interactive
// demo mode. It cycles through predefined choices and texts.
type demoPrompter struct {
	choices   []int
	texts     []string
	choiceIdx int
	textIdx   int
}

func (d *demoPrompter) Ask(
	question string,
	options []string,
) (int, error) {
	if d.choiceIdx >= len(d.choices) {
		return 0, errors.New("demo: no more choices")
	}
	choice := d.choices[d.choiceIdx]
	d.choiceIdx++

	fmt.Println(cli.RenderQuestion(question))
	if choice < len(options) {
		fmt.Println(cli.RenderAnswer(options[choice]))
	}
	fmt.Println()

	return choice, nil
}

func (d *demoPrompter) AskText(
	question string,
) (string, error) {
	if d.textIdx >= len(d.texts) {
		return "", errors.New("demo: no more texts")
	}
	text := d.texts[d.textIdx]
	d.textIdx++

	fmt.Println(cli.RenderQuestion(question))
	fmt.Println(cli.RenderAnswer(text))
	fmt.Println()

	return text, nil
}

func main() {
	ctx := context.Background()

	// Parse flags.
	tutorialsDir := consts.DefaultTutorialsDir
	demoMode := false

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--tutorials":
			if i+1 < len(os.Args) {
				tutorialsDir = os.Args[i+1]
				i++
			}
		case "--demo":
			demoMode = true
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Error: %v\n", err,
		)
		os.Exit(1)
	}

	cachePath := filepath.Join(
		homeDir,
		".config",
		consts.CacheDir,
		consts.ReleaseCacheFile,
	)

	// Ensure cache directory exists.
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Warning: could not create cache "+
				"dir: %v\n",
			err,
		)
	}

	configPath := filepath.Join(
		".", consts.OpenCodeConfigFile,
	)

	httpClient := http.DefaultClient
	fetcher := func(
		fetchCtx context.Context,
	) ([]schema.Release, error) {
		return schema.FetchReleases(
			fetchCtx, httpClient,
		)
	}

	var cfg *cli.SetupConfig

	if demoMode {
		// Demo mode: skip MCP setup and version
		// selection, go straight to role discovery
		// with predefined inputs.
		demo := &demoPrompter{
			choices: []int{
				2, // Skip MCP installation
				0, // Security Engineer
			},
			texts: []string{
				"CI/CD pipeline management, " +
					"dependency management, " +
					"and coding with upstream " +
					"open-source components",
			},
		}

		cfg = &cli.SetupConfig{
			Prompter:      demo,
			BinaryLookup:  mcp.DefaultBinaryLookup,
			PodmanChecker: mcp.DefaultPodmanChecker,
			Installer: mcp.NewInstaller(
				mcp.DefaultReleaseFetcher,
				mcp.DefaultCommandRunner,
			),
			ConfigPath:   configPath,
			RolePrompter: demo,
			TutorialsDir: tutorialsDir,
		}
	} else {
		prompter := &huhPrompter{}
		cfg = &cli.SetupConfig{
			Prompter:      prompter,
			BinaryLookup:  mcp.DefaultBinaryLookup,
			PodmanChecker: mcp.DefaultPodmanChecker,
			Installer: mcp.NewInstaller(
				mcp.DefaultReleaseFetcher,
				mcp.DefaultCommandRunner,
			),
			SSHChecker:       mcp.DefaultSSHChecker,
			ConfigPath:       configPath,
			VersionFetcher:   fetcher,
			VersionCachePath: cachePath,
			RolePrompter:     prompter,
			TutorialsDir:     tutorialsDir,
		}
	}

	lipgloss.Println(cli.RenderBanner())

	result, err := cli.RunSetup(ctx, cfg, os.Stdout)
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "\nSetup error: %v\n", err,
		)
		os.Exit(1)
	}

	lipgloss.Println(cli.RenderSessionStatus(
		result.Session.SchemaVersion,
		result.Session.IsFallback(),
	))

	// Display role info if available.
	if result.Session.GetRoleName() != "" {
		lipgloss.Println(cli.RenderSessionRoleInfo(
			result.Session.GetRoleName(),
			result.Session.LearningPathSteps,
		))
	}

	if demoMode {
		// Demo mode: run team and authoring flows
		// with predefined inputs.
		runDemoTeamAndAuthoring(
			result.Session, tutorialsDir,
		)
		return
	}

	// Interactive mode: present main menu.
	prompter := cfg.RolePrompter
	if prompter == nil {
		prompter = &huhPrompter{}
	}
	runMainMenu(
		ctx, prompter, result.Session,
		tutorialsDir, cfg,
	)
}

// runMainMenu presents the interactive main menu after
// setup completes. Users can start tutorials, configure
// teams, author artifacts, or exit.
func runMainMenu(
	ctx context.Context,
	prompter cli.FreeTextPrompter,
	sess *session.Session,
	tutorialsDir string,
	cfg *cli.SetupConfig,
) {
	for {
		fmt.Println()
		choice, err := prompter.Ask(
			"What would you like to do?",
			[]string{
				"Start a tutorial",
				"Configure team collaboration",
				"Author a Gemara artifact",
				"Update MCP server",
				"Exit",
			},
		)
		if err != nil {
			fmt.Fprintf(
				os.Stderr, "\nError: %v\n", err,
			)
			return
		}

		switch choice {
		case 0:
			runTutorialFlow(prompter, sess, tutorialsDir)
		case 1:
			runTeamFlow(prompter, sess, tutorialsDir)
		case 2:
			runAuthoringFlow(
				prompter, sess, tutorialsDir,
			)
		case 3:
			runMCPUpdate(ctx, cfg, sess)
		default:
			fmt.Println()
			fmt.Println(cli.RenderSuccess(
				"Session complete. Goodbye!",
			))
			return
		}
	}
}

// runTutorialFlow displays the user's learning path and
// offers to start a tutorial step.
func runTutorialFlow(
	prompter cli.FreeTextPrompter,
	sess *session.Session,
	tutorialsDir string,
) {
	if sess.GetRoleName() == "" {
		fmt.Println(cli.RenderWarning(
			"No role configured. Run setup first.",
		))
		return
	}

	if sess.LearningPathSteps == 0 {
		fmt.Println(cli.RenderNote(
			"No tutorials matched your role and " +
				"activities. Try different keywords.",
		))
		return
	}

	lipgloss.Println(cli.RenderSessionRoleInfo(
		sess.GetRoleName(),
		sess.LearningPathSteps,
	))

	fmt.Println()
	fmt.Println(cli.RenderNote(
		"Tutorial content is delivered through the " +
			"Gemara tutorials in " + tutorialsDir + ". " +
			"Follow the learning path steps above " +
			"in order.",
	))
}

// runTeamFlow runs the interactive team collaboration
// setup.
func runTeamFlow(
	prompter cli.FreeTextPrompter,
	sess *session.Session,
	tutorialsDir string,
) {
	teamCfg := &cli.TeamPromptConfig{
		Prompter:      prompter,
		TutorialsDir:  tutorialsDir,
		SchemaVersion: sess.SchemaVersion,
		TeamConfigDir: filepath.Join(
			homeConfigDir(), consts.MCPInstallDir,
			"teams",
		),
	}

	teamResult, err := cli.RunTeamSetup(
		teamCfg, os.Stdout,
	)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"\nTeam setup error: %v\n", err,
		)
		return
	}

	if teamResult.Team != nil {
		sess.SetTeamInfo(
			teamResult.Team.Name,
			len(teamResult.Team.Members),
		)
	}
}

// runAuthoringFlow runs the interactive guided authoring
// flow.
func runAuthoringFlow(
	prompter cli.FreeTextPrompter,
	sess *session.Session,
	tutorialsDir string,
) {
	outputDir := filepath.Join(
		".", consts.AuthoringOutputDir,
	)

	authorCfg := &cli.AuthorPromptConfig{
		Prompter:      prompter,
		Session:       sess,
		SchemaVersion: sess.SchemaVersion,
		OutputDir:     outputDir,
		OutputFormat:  consts.DefaultArtifactFormat,
		RoleName:      sess.GetRoleName(),
		Keywords:      sess.ActivityKeywords,
	}

	_, err := cli.RunGuidedAuthoring(
		authorCfg, os.Stdout,
	)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"\nAuthoring error: %v\n", err,
		)
	}
}

// runMCPUpdate checks for and offers to install MCP server
// updates.
func runMCPUpdate(
	ctx context.Context,
	cfg *cli.SetupConfig,
	sess *session.Session,
) {
	if cfg.Installer == nil {
		fmt.Println(cli.RenderWarning(
			"MCP installer not available.",
		))
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(cli.RenderWarning(
			"Could not determine home directory.",
		))
		return
	}
	installDir := filepath.Join(
		homeDir, ".local", "share", consts.MCPInstallDir,
	)

	update, err := cfg.Installer.CheckForUpdate(
		ctx, installDir,
	)
	if err != nil {
		fmt.Println(cli.RenderWarning(
			"Update check failed: " + err.Error(),
		))
		return
	}

	if !update.UpdateAvailable {
		if update.Installed != nil {
			fmt.Println(cli.RenderSuccess(fmt.Sprintf(
				"MCP server is up to date (%s, "+
					"commit %s)",
				update.Installed.Tag,
				truncateSHAStr(
					update.Installed.CommitSHA,
				),
			)))
		} else {
			fmt.Println(cli.RenderNote(
				"No installed MCP server found. " +
					"Use 'Build from source' during " +
					"setup to install.",
			))
		}
		return
	}

	fmt.Println()
	fmt.Println(cli.RenderStatus(fmt.Sprintf(
		"Update available: %s (commit %s) -> "+
			"%s (commit %s)",
		update.Installed.Tag,
		truncateSHAStr(update.Installed.CommitSHA),
		update.Latest.Tag,
		truncateSHAStr(update.Latest.CommitSHA),
	)))

	choice, err := cfg.Prompter.Ask(
		"Update gemara-mcp?",
		[]string{"Yes, update now", "Skip"},
	)
	if err != nil || choice != 0 {
		return
	}

	// Perform update via the setup flow's update logic.
	fmt.Println(cli.RenderStatus(
		"Updating MCP server...",
	))
	// Re-use checkAndOfferUpdate indirectly — just
	// trigger a fresh build at the latest SHA.
	method := mcp.CloneHTTPS
	if cfg.SSHChecker != nil {
		method = mcp.DetectCloneMethod(
			ctx, cfg.SSHChecker,
		)
	}
	binaryPath, buildErr := cfg.Installer.CloneAndBuild(
		ctx, method, update.Latest, installDir,
	)
	if buildErr != nil {
		fmt.Println(cli.RenderWarning(
			"Update failed: " + buildErr.Error(),
		))
		return
	}

	// Save metadata.
	installed := &mcp.InstalledRelease{
		Tag:        update.Latest.Tag,
		CommitSHA:  update.Latest.CommitSHA,
		Prerelease: update.Latest.Prerelease,
		InstalledAt: time.Now().UTC().Format(
			time.RFC3339,
		),
		BinaryPath: binaryPath,
	}
	_ = mcp.SaveInstalledRelease(installDir, installed)

	// Update MCP config.
	if cfg.ConfigPath != "" {
		config, readErr := mcp.ReadOpenCodeConfig(
			cfg.ConfigPath,
		)
		if readErr == nil {
			mcp.EnsureMCPEntry(config, binaryPath)
			_ = mcp.WriteOpenCodeConfig(
				cfg.ConfigPath, config,
			)
		}
	}

	fmt.Println(cli.RenderSuccess(fmt.Sprintf(
		"Updated to %s (commit %s)",
		update.Latest.Tag,
		truncateSHAStr(update.Latest.CommitSHA),
	)))
}

// homeConfigDir returns a safe config directory path.
func homeConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config")
}

// truncateSHAStr truncates a SHA to 12 chars for display.
func truncateSHAStr(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

// runDemoTeamAndAuthoring runs the US5 team collaboration
// view and US6 guided authoring flow in demo mode with
// predefined inputs.
func runDemoTeamAndAuthoring(
	sess *session.Session,
	tutorialsDir string,
) {
	// --- US5: Team Collaboration View ---
	//
	// Team prompt call sequence:
	// AskText: team name
	// AskText: member 1 name
	// Ask:     member 1 role (0=Security Engineer)
	// AskText: member 2 name
	// Ask:     member 2 role (1=Compliance Officer)
	// AskText: member 3 name
	// Ask:     member 3 role (3=Developer)
	// AskText: "" (finish adding members)
	teamDemo := &demoPrompter{
		choices: []int{0, 1, 3},
		texts: []string{
			"GRC Product Team",
			"Alice",
			"Bob",
			"Carol",
			"", // finish adding members
		},
	}

	teamCfg := &cli.TeamPromptConfig{
		Prompter:      teamDemo,
		TutorialsDir:  tutorialsDir,
		SchemaVersion: sess.SchemaVersion,
		TeamConfigDir: filepath.Join(
			os.TempDir(), "pacman-demo-teams",
		),
	}

	teamResult, err := cli.RunTeamSetup(
		teamCfg, os.Stdout,
	)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"\nTeam setup error: %v\n", err,
		)
		return
	}

	if teamResult.Team != nil {
		sess.SetTeamInfo(
			teamResult.Team.Name,
			len(teamResult.Team.Members),
		)
	}

	// --- US6: Guided Gemara Content Authoring ---
	//
	// Authoring flow call sequence:
	// Ask:     artifact type (2=ThreatCatalog)
	//
	// Step 1 — metadata (3 fields):
	// AskText: name (required)
	// AskText: description (required)
	// AskText: version (optional)
	//
	// Step 2 — scope (2 fields):
	// AskText: scope (required)
	// AskText: boundary (optional)
	//
	// Step 3 — capabilities (2 fields):
	// AskText: capability_name (required)
	// AskText: capability_description (required)
	//
	// Step 4 — threats (3 fields):
	// AskText: threat_id (required)
	// AskText: threat_description (required)
	// AskText: target_capability (required)
	authorDemo := &demoPrompter{
		choices: []int{2}, // ThreatCatalog
		texts: []string{
			// metadata
			"ACME.WEB.THR01",
			"Threat catalog for web application " +
				"attack surface assessment",
			"1.0.0",
			// scope
			"Web application authentication " +
				"and session management",
			"Third-party SaaS integrations",
			// capabilities
			"Authentication",
			"User identity verification and " +
				"session management",
			// threats
			"THR-001",
			"SQL injection via unvalidated " +
				"user input in search parameters",
			"Authentication",
		},
	}

	outputDir := filepath.Join(
		os.TempDir(), "pacman-demo-artifacts",
	)

	authorCfg := &cli.AuthorPromptConfig{
		Prompter:      authorDemo,
		Session:       sess,
		SchemaVersion: sess.SchemaVersion,
		OutputDir:     outputDir,
		OutputFormat:  consts.DefaultArtifactFormat,
		RoleName:      sess.GetRoleName(),
		Keywords:      sess.ActivityKeywords,
	}

	authorResult, err := cli.RunGuidedAuthoring(
		authorCfg, os.Stdout,
	)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"\nAuthoring error: %v\n", err,
		)
		return
	}

	if authorResult.Artifact != nil {
		fmt.Println()
		lipgloss.Println(cli.RenderSessionStatus(
			sess.SchemaVersion,
			sess.IsFallback(),
		))
		if sess.GetRoleName() != "" {
			lipgloss.Println(cli.RenderSessionRoleInfo(
				sess.GetRoleName(),
				sess.LearningPathSteps,
			))
		}
	}
}
