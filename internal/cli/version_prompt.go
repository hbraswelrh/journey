// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/hbraswelrh/pacman/internal/mcp"
	"github.com/hbraswelrh/pacman/internal/schema"
	"github.com/hbraswelrh/pacman/internal/session"
)

// ReleaseFetcherFn fetches releases from upstream.
type ReleaseFetcherFn func(
	ctx context.Context,
) ([]schema.Release, error)

// VersionPromptConfig holds dependencies for the version
// selection flow.
type VersionPromptConfig struct {
	// Prompter handles user interaction.
	Prompter UserPrompter
	// Fetcher retrieves releases from upstream.
	Fetcher ReleaseFetcherFn
	// CachePath is the path to the local release cache.
	CachePath string
	// Session is the current session to update.
	Session *session.Session
	// MCPClient is the MCP client for compatibility checks.
	// May be nil if MCP is not installed.
	MCPClient *mcp.Client
}

// RunVersionSelection executes the schema version selection
// flow:
//  1. Fetch or load cached releases.
//  2. Determine Stable and Latest versions.
//  3. Prompt the user to choose.
//  4. Apply selection to the session.
//  5. Display warnings for experimental schemas or version
//     mismatches.
func RunVersionSelection(
	ctx context.Context,
	cfg *VersionPromptConfig,
	out io.Writer,
) error {
	// Step 1: Fetch or cache releases.
	cached, err := schema.RefreshOrCache(
		ctx, cfg.Fetcher, cfg.CachePath,
	)
	if err != nil {
		return fmt.Errorf(
			"fetch schema versions: %w", err,
		)
	}

	if cached.FromCache {
		fmt.Fprintf(
			out,
			"Using cached version data (fetched %s). "+
				"Upstream was not reachable.\n",
			cached.LastFetched.Format("2006-01-02"),
		)
	}

	// Step 2: Determine versions.
	choice, err := schema.DetermineVersions(
		cached.Releases,
	)
	if err != nil {
		return fmt.Errorf("determine versions: %w", err)
	}

	// Step 3: Build prompt options.
	options, selectionMap := buildOptions(choice)

	fmt.Fprintf(out, "\nGemara Schema Version Selection\n")
	fmt.Fprintf(out, "-------------------------------\n")

	if choice.StableVersion != nil {
		fmt.Fprintf(
			out,
			"Stable: %s — all core schemas are Stable\n",
			choice.StableVersion.Tag,
		)
	}
	if choice.LatestVersion != nil {
		fmt.Fprintf(
			out,
			"Latest: %s",
			choice.LatestVersion.Tag,
		)
		expSchemas := experimentalNames(
			choice.LatestSchemaStatus,
		)
		if len(expSchemas) > 0 {
			fmt.Fprintf(
				out,
				" — Experimental schemas: %v",
				expSchemas,
			)
		}
		fmt.Fprintf(out, "\n")
	}

	fmt.Fprintf(out, "\n")

	idx, err := cfg.Prompter.Ask(
		"Select a schema version:", options,
	)
	if err != nil {
		return fmt.Errorf("prompt version: %w", err)
	}

	selection := selectionMap[idx]

	// Step 4: Apply selection.
	result, err := schema.SelectVersion(
		choice,
		selection,
		cfg.Session,
		cfg.MCPClient,
		nil, // No confirmer for initial selection.
	)
	if err != nil {
		return fmt.Errorf("select version: %w", err)
	}

	fmt.Fprintf(
		out,
		"Selected schema version: %s\n",
		result.SelectedTag,
	)

	// Step 5: Warnings.
	if len(result.ExperimentalSchemas) > 0 {
		fmt.Fprintf(
			out,
			"Note: the following schemas are "+
				"Experimental at %s: %v\n",
			result.SelectedTag,
			result.ExperimentalSchemas,
		)
	}

	if result.CompatWarning != "" {
		fmt.Fprintf(
			out,
			"Warning: %s\n",
			result.CompatWarning,
		)
	}

	// Newer version notification (when user picks Stable
	// and a newer Latest exists).
	if selection == schema.SelectionStable &&
		choice.LatestVersion != nil &&
		choice.StableVersion != nil &&
		choice.LatestVersion.Tag !=
			choice.StableVersion.Tag {
		fmt.Fprintf(
			out,
			"Note: a newer version (%s) is available "+
				"upstream but contains Experimental "+
				"schemas.\n",
			choice.LatestVersion.Tag,
		)
	}

	return nil
}

// buildOptions creates the prompt option list and a mapping
// from option index to SelectionType.
func buildOptions(
	choice *schema.VersionChoice,
) ([]string, map[int]schema.SelectionType) {
	var options []string
	selMap := make(map[int]schema.SelectionType)

	idx := 0
	if choice.StableVersion != nil {
		options = append(
			options,
			fmt.Sprintf(
				"Stable (%s)", choice.StableVersion.Tag,
			),
		)
		selMap[idx] = schema.SelectionStable
		idx++
	}
	if choice.LatestVersion != nil {
		options = append(
			options,
			fmt.Sprintf(
				"Latest (%s)", choice.LatestVersion.Tag,
			),
		)
		selMap[idx] = schema.SelectionLatest
	}

	return options, selMap
}

// experimentalNames returns schema names marked Experimental.
func experimentalNames(
	statusMap map[string]schema.SchemaStatus,
) []string {
	var names []string
	for name, status := range statusMap {
		if status == schema.StatusExperimental {
			names = append(names, name)
		}
	}
	return names
}
