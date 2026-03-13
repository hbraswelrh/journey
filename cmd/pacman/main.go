// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
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

// huhPrompter implements cli.UserPrompter using the
// charmbracelet/huh interactive select widget.
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

func main() {
	ctx := context.Background()

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

	cfg := &cli.SetupConfig{
		Prompter:      &huhPrompter{},
		BinaryLookup:  mcp.DefaultBinaryLookup,
		PodmanChecker: mcp.DefaultPodmanChecker,
		Installer: mcp.NewInstaller(
			mcp.DefaultReleaseFetcher,
			mcp.DefaultCommandRunner,
		),
		ConfigPath:       configPath,
		VersionFetcher:   fetcher,
		VersionCachePath: cachePath,
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
}
