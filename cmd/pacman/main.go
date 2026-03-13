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
}
