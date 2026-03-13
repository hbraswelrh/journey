// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

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

	// Demo mode: continue with team collaboration and
	// guided authoring demonstrations.
	if demoMode {
		runDemoTeamAndAuthoring(
			result.Session, tutorialsDir,
		)
	}
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
