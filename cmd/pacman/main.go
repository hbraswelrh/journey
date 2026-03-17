// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"charm.land/huh/v2"
	lipgloss "charm.land/lipgloss/v2"
	"golang.org/x/term"

	"github.com/hbraswelrh/pacman/internal/cli"
	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/mcp"
	"github.com/hbraswelrh/pacman/internal/roles"
	"github.com/hbraswelrh/pacman/internal/schema"
	"github.com/hbraswelrh/pacman/internal/session"
	"github.com/hbraswelrh/pacman/internal/tutorials"
)

// isInteractiveTTY returns true if both stdin and stdout
// are connected to a terminal. When running inside
// OpenCode or a pipe, this returns false.
func isInteractiveTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) &&
		term.IsTerminal(int(os.Stdout.Fd()))
}

// plainPrompter implements cli.FreeTextPrompter using
// simple numbered menus and line-based input. Used when
// no TTY is detected (e.g., running inside OpenCode).
type plainPrompter struct {
	reader *bufio.Reader
}

func newPlainPrompter() *plainPrompter {
	return &plainPrompter{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (p *plainPrompter) Ask(
	question string,
	options []string,
) (int, error) {
	fmt.Println()
	fmt.Println(question)
	for i, opt := range options {
		fmt.Printf("  [%d] %s\n", i+1, opt)
	}
	fmt.Print("Enter number: ")

	line, err := p.reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("read input: %w", err)
	}
	line = strings.TrimSpace(line)
	n, err := strconv.Atoi(line)
	if err != nil || n < 1 || n > len(options) {
		return 0, fmt.Errorf(
			"invalid choice: %q (enter 1-%d)",
			line, len(options),
		)
	}
	return n - 1, nil
}

func (p *plainPrompter) AskText(
	question string,
) (string, error) {
	fmt.Println()
	fmt.Print(question + " ")

	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	return strings.TrimSpace(line), nil
}

func (p *plainPrompter) AskMultiSelect(
	question string,
	options []string,
	defaults []int,
) ([]int, error) {
	fmt.Println()
	fmt.Println(question)
	for i, opt := range options {
		marker := " "
		for _, d := range defaults {
			if d == i {
				marker = "*"
				break
			}
		}
		fmt.Printf("  [%s] %d. %s\n", marker, i+1, opt)
	}
	fmt.Print(
		"Enter numbers separated by commas " +
			"(or press Enter for defaults): ",
	)

	line, err := p.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaults, nil
	}

	var selected []int
	for _, part := range strings.Split(line, ",") {
		part = strings.TrimSpace(part)
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 || n > len(options) {
			continue
		}
		selected = append(selected, n-1)
	}
	return selected, nil
}

func (p *plainPrompter) AskConfirm(
	question string,
) (bool, error) {
	fmt.Println()
	fmt.Print(question + " (y/n): ")

	line, err := p.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read input: %w", err)
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes", nil
}

func (p *plainPrompter) AskTextWithDefault(
	question string,
	defaultValue string,
) (string, error) {
	fmt.Println()
	prompt := question
	if defaultValue != "" {
		prompt += " [" + defaultValue + "]"
	}
	fmt.Print(prompt + ": ")

	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue, nil
	}
	return line, nil
}

// huhPrompter implements cli.FreeTextPrompter using the
// charmbracelet/huh interactive widgets with the Charm
// theme for a polished appearance. Requires a TTY.
type huhPrompter struct{}

// pacmanTheme wraps the Charm theme for consistent styling
// across all interactive widgets.
var pacmanTheme huh.Theme = huh.ThemeFunc(huh.ThemeCharm)

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
		WithTheme(pacmanTheme).
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
		WithTheme(pacmanTheme).
		Run()
	if err != nil {
		return "", fmt.Errorf("prompt: %w", err)
	}

	return answer, nil
}

func (p *huhPrompter) AskMultiSelect(
	question string,
	options []string,
	defaults []int,
) ([]int, error) {
	var selected []int

	opts := make([]huh.Option[int], len(options))
	for i, label := range options {
		opts[i] = huh.NewOption(label, i)
	}

	ms := huh.NewMultiSelect[int]().
		Title(question).
		Options(opts...).
		Value(&selected)

	if len(defaults) > 0 {
		ms.Value(&defaults)
	}

	if err := ms.WithTheme(pacmanTheme).Run(); err != nil {
		return nil, fmt.Errorf("prompt: %w", err)
	}

	if len(defaults) > 0 && len(selected) == 0 {
		return defaults, nil
	}
	return selected, nil
}

func (p *huhPrompter) AskConfirm(
	question string,
) (bool, error) {
	var confirmed bool

	err := huh.NewConfirm().
		Title(question).
		Value(&confirmed).
		WithTheme(pacmanTheme).
		Run()
	if err != nil {
		return false, fmt.Errorf("prompt: %w", err)
	}

	return confirmed, nil
}

func (p *huhPrompter) AskTextWithDefault(
	question string,
	defaultValue string,
) (string, error) {
	answer := defaultValue

	err := huh.NewInput().
		Title(question).
		Value(&answer).
		WithTheme(pacmanTheme).
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

func (d *demoPrompter) AskMultiSelect(
	question string,
	options []string,
	defaults []int,
) ([]int, error) {
	fmt.Println(cli.RenderQuestion(question))
	// In demo mode, accept all defaults.
	selected := defaults
	if len(selected) == 0 && len(options) > 0 {
		// Select first option if no defaults.
		selected = []int{0}
	}
	for _, idx := range selected {
		if idx < len(options) {
			fmt.Println(cli.RenderAnswer(
				"  [x] " + options[idx],
			))
		}
	}
	fmt.Println()
	return selected, nil
}

func (d *demoPrompter) AskConfirm(
	question string,
) (bool, error) {
	fmt.Println(cli.RenderQuestion(question))
	fmt.Println(cli.RenderAnswer("Yes"))
	fmt.Println()
	return true, nil
}

func (d *demoPrompter) AskTextWithDefault(
	question string,
	defaultValue string,
) (string, error) {
	fmt.Println(cli.RenderQuestion(question))
	fmt.Println(cli.RenderAnswer(defaultValue))
	fmt.Println()
	return defaultValue, nil
}

func main() {
	// Parse flags.
	doctorMode := false
	helpMode := false

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--doctor":
			doctorMode = true
		case "--help", "-h":
			helpMode = true
		}
	}

	if helpMode {
		printUsage()
		return
	}

	configPath := filepath.Join(
		".", consts.OpenCodeConfigFile,
	)

	if doctorMode {
		doctorCfg := cli.DefaultDoctorConfig(configPath)
		ok := cli.RunDoctor(doctorCfg, os.Stdout)
		if !ok {
			os.Exit(1)
		}
		return
	}

	// Default: show usage.
	printUsage()
}

func printUsage() {
	lipgloss.Println(cli.RenderBanner())
	fmt.Println(
		"Pac-Man verifies your environment for " +
			"Gemara tutorials.",
	)
	fmt.Println(
		"The tutorial experience is delivered " +
			"through OpenCode.",
	)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(
		"  ./pacman --doctor    " +
			"Check environment readiness",
	)
	fmt.Println(
		"  ./pacman --help      " +
			"Show this help message",
	)
	fmt.Println()
	fmt.Println("Getting started:")
	fmt.Println(
		"  1. Run ./pacman --doctor to verify " +
			"your setup",
	)
	fmt.Println(
		"  2. Start OpenCode: opencode",
	)
	fmt.Println(
		"  3. Tell OpenCode your role and what " +
			"you want to do",
	)
	fmt.Println()
	fmt.Println("Example prompts for OpenCode:")
	fmt.Println(
		"  \"I'm a Security Engineer working on " +
			"CI/CD pipeline security.\"",
	)
	fmt.Println(
		"  \"I'm a Policy Author and need to " +
			"create an adherence policy.\"",
	)
	fmt.Println(
		"  \"Run the threat_assessment prompt " +
			"for my web application.\"",
	)
	fmt.Println()
}

// --- Legacy interactive functions below ---
// These are retained for the test suite and may be removed
// in a future cleanup. The runtime entry point (main)
// only uses --doctor and --help.

// legacyUnused suppresses "unused" lint warnings for
// variables referenced only by the retained legacy
// functions.
var _ = func() {
	_ = context.Background
	_ = http.DefaultClient
	_ = schema.FetchReleases
	_ = tutorials.ExpandTutorialsDir
}

func legacyMain() { //nolint:unused
	ctx := context.Background()
	tutorialsDir := consts.DefaultTutorialsDir
	demoMode := false
	doctorMode := false
	forceInteractive := false
	_ = demoMode
	_ = doctorMode
	_ = forceInteractive
	_ = ctx

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Error: %v\n", err,
		)
		os.Exit(1)
	}

	// Resolve tutorials directory — expand ~ and check
	// managed clone location.
	tutorialsDir = tutorials.ExpandTutorialsDir(
		tutorialsDir, homeDir,
	)

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

	// Doctor mode: check environment and exit.
	if doctorMode {
		doctorCfg := cli.DefaultDoctorConfig(configPath)
		ok := cli.RunDoctor(doctorCfg, os.Stdout)
		if !ok {
			os.Exit(1)
		}
		return
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
		// Detect whether we have a TTY for interactive
		// huh widgets, or fall back to plain prompts
		// (e.g., when running inside OpenCode).
		var prompter cli.FreeTextPrompter
		if forceInteractive || isInteractiveTTY() {
			prompter = &huhPrompter{}
		} else {
			fmt.Println(cli.RenderNote(
				"No interactive terminal detected. " +
					"Using simple text prompts. " +
					"For the full interactive " +
					"experience, run ./pacman " +
					"directly in a terminal.",
			))
			prompter = newPlainPrompter()
		}
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
			result.Session.RecommendedArtifacts,
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
				"  Start a tutorial",
				"  Launch a wizard (MCP-assisted)",
				"  Configure team collaboration",
				"  Author a Gemara artifact",
				"  Update MCP server",
				"  Check environment (doctor)",
				"  Exit",
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
			runWizardFlow(prompter, sess)
		case 2:
			runTeamFlow(prompter, sess, tutorialsDir)
		case 3:
			runAuthoringFlow(
				prompter, sess, tutorialsDir,
			)
		case 4:
			runMCPUpdate(ctx, cfg, sess)
		case 5:
			doctorCfg := cli.DefaultDoctorConfig(
				cfg.ConfigPath,
			)
			cli.RunDoctor(doctorCfg, os.Stdout)
		default:
			fmt.Println()
			fmt.Println(cli.RenderSuccess(
				"Session complete. Goodbye!",
			))
			return
		}
	}
}

// runTutorialFlow regenerates the learning path and
// launches the interactive tutorial player.
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

	// Load tutorials — fetch from upstream if not found.
	tuts, err := tutorials.LoadTutorials(tutorialsDir)
	if err != nil {
		fmt.Println(cli.RenderNote(
			"Tutorials not found locally. " +
				"Fetching from upstream Gemara " +
				"repository...",
		))
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			fmt.Println(cli.RenderWarning(
				"Could not determine home " +
					"directory: " + homeErr.Error(),
			))
			return
		}
		fetchResult, fetchErr := tutorials.ResolveTutorialsDir(
			&tutorials.FetchConfig{
				HomeDir: homeDir,
			},
		)
		if fetchErr != nil {
			fmt.Println(cli.RenderWarning(
				"Could not fetch tutorials: " +
					fetchErr.Error(),
			))
			return
		}
		if fetchResult.Cloned {
			fmt.Println(cli.RenderSuccess(
				"Cloned Gemara repository to " +
					fetchResult.TutorialsDir,
			))
		} else if fetchResult.Updated {
			fmt.Println(cli.RenderSuccess(
				"Updated tutorials from upstream",
			))
		}
		tutorialsDir = fetchResult.TutorialsDir
		tuts, err = tutorials.LoadTutorials(
			tutorialsDir,
		)
		if err != nil {
			fmt.Println(cli.RenderWarning(
				"Could not load tutorials after " +
					"fetch: " + err.Error(),
			))
			return
		}
	}

	// Build a minimal activity profile for path
	// generation.
	profile := buildProfileFromSession(sess)
	path := tutorials.GeneratePath(
		profile, tuts, sess.SchemaVersion,
	)

	if path == nil || len(path.Steps) == 0 {
		fmt.Println(cli.RenderNote(
			"No tutorials matched your role and " +
				"activities. Try different keywords.",
		))
		return
	}

	// Launch the interactive tutorial player.
	tutCfg := &cli.TutorialPromptConfig{
		Prompter:     prompter,
		LearningPath: path,
		TutorialsDir: tutorialsDir,
		RoleName:     sess.GetRoleName(),
		Keywords:     sess.ActivityKeywords,
		Session:      sess,
	}

	result, err := cli.RunTutorialPlayer(
		tutCfg, os.Stdout,
	)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"\nTutorial error: %v\n", err,
		)
		return
	}

	// Update session with completion tracking.
	completed := 0
	for idx, done := range result.CompletedSteps {
		if done {
			completed++
			if idx < len(path.Steps) {
				sess.MarkTutorialComplete(
					path.Steps[idx].Tutorial.Title,
				)
			}
		}
	}
	if completed > 0 {
		fmt.Println(cli.RenderSuccess(fmt.Sprintf(
			"%d of %d tutorials completed this session",
			completed, len(path.Steps),
		)))
	}
}

// runWizardFlow launches the MCP wizard selector and
// argument collector.
func runWizardFlow(
	prompter cli.FreeTextPrompter,
	sess *session.Session,
) {
	// Check if the prompter supports wizard operations
	// (multi-select, confirm, text-with-default).
	wizPrompter, ok := prompter.(cli.WizardPrompter)
	if !ok {
		fmt.Println(cli.RenderWarning(
			"Wizard flow requires an interactive " +
				"terminal with multi-select support.",
		))
		return
	}

	wizCfg := &cli.WizardPromptConfig{
		Prompter:     wizPrompter,
		MCPAvailable: !sess.IsFallback(),
		ServerMode:   sess.GetServerMode(),
		RoleName:     sess.GetRoleName(),
	}

	_, err := cli.RunWizardLauncher(wizCfg, os.Stdout)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"\nWizard error: %v\n", err,
		)
	}
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

	// Update MCP config, preserving existing mode.
	if cfg.ConfigPath != "" {
		config, readErr := mcp.ReadOpenCodeConfig(
			cfg.ConfigPath,
		)
		if readErr == nil {
			mode := consts.MCPModeDefault
			if entry, ok := config.MCP[consts.MCPServerName]; ok {
				mode = mcp.ParseMCPMode(entry)
			}
			mcp.EnsureMCPEntry(
				config, binaryPath, mode,
			)
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

// buildProfileFromSession creates a minimal ActivityProfile
// from session state for learning path regeneration.
func buildProfileFromSession(
	sess *session.Session,
) *roles.ActivityProfile {
	var layers []roles.LayerMapping
	for _, l := range sess.ResolvedLayers {
		layers = append(layers, roles.LayerMapping{
			Layer:      l,
			Confidence: roles.ConfidenceStrong,
		})
	}

	predefined := roles.PredefinedRoles()
	var role *roles.Role
	for i := range predefined {
		if predefined[i].Name == sess.GetRoleName() {
			role = &predefined[i]
			break
		}
	}

	return &roles.ActivityProfile{
		Role:              role,
		ExtractedKeywords: sess.ActivityKeywords,
		ResolvedLayers:    layers,
	}
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
				sess.RecommendedArtifacts,
			))
		}
	}
}
