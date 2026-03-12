// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/hbraswelrh/pacman/internal/mcp"
	"github.com/hbraswelrh/pacman/internal/session"
)

// SelectionType represents the user's version selection.
type SelectionType int

const (
	// SelectionStable selects the most recent version where
	// all core schemas are Stable.
	SelectionStable SelectionType = iota
	// SelectionLatest selects the most recent version tag
	// regardless of schema stability.
	SelectionLatest
)

// VersionChoice holds the Stable and Latest version options
// determined from a release list, along with per-schema
// status maps for display.
type VersionChoice struct {
	// StableVersion is the most recent release where all
	// core schemas are Stable. May be nil if no release
	// qualifies.
	StableVersion *Release
	// LatestVersion is the most recent release by date.
	// May be nil if the release list is empty.
	LatestVersion *Release
	// StableSchemaStatus maps schema names to their status
	// at the Stable version.
	StableSchemaStatus map[string]SchemaStatus
	// LatestSchemaStatus maps schema names to their status
	// at the Latest version.
	LatestSchemaStatus map[string]SchemaStatus
}

// SelectionResult holds the outcome of a version selection.
type SelectionResult struct {
	// SelectedTag is the version tag that was selected.
	SelectedTag string
	// ExperimentalSchemas lists schema names that are
	// Experimental at the selected version.
	ExperimentalSchemas []string
	// CompatWarning is a warning message when the MCP
	// server version does not match the selected schema
	// version. Empty if no MCP client or versions match.
	CompatWarning string
	// PreviousVersion is the version that was replaced in
	// a mid-session switch. Empty for initial selection.
	PreviousVersion string
}

// ErrVersionSwitchDeclined is returned when the user declines
// a mid-session version switch.
var ErrVersionSwitchDeclined = errors.New(
	"version switch declined by user",
)

// ErrNoVersionAvailable is returned when neither Stable nor
// Latest can be determined from the release list.
var ErrNoVersionAvailable = errors.New(
	"no version available in release list",
)

// VersionSwitchConfirmer asks the user to confirm a
// mid-session version switch. It receives the current and
// proposed version tags and returns true if the user confirms.
type VersionSwitchConfirmer func(
	currentVersion string,
	proposedVersion string,
) (bool, error)

// DetermineVersions analyzes a release list and identifies
// the Stable and Latest version options with their schema
// status maps.
func DetermineVersions(
	releases []Release,
) (*VersionChoice, error) {
	if len(releases) == 0 {
		return nil, ErrNoVersionAvailable
	}

	stable := DetermineStableVersion(releases)
	latest := DetermineLatestVersion(releases)

	choice := &VersionChoice{
		StableVersion:      stable,
		LatestVersion:      latest,
		StableSchemaStatus: make(map[string]SchemaStatus),
		LatestSchemaStatus: make(map[string]SchemaStatus),
	}

	if stable != nil {
		for k, v := range stable.SchemaStatusMap {
			choice.StableSchemaStatus[k] = v
		}
	}

	if latest != nil {
		for k, v := range latest.SchemaStatusMap {
			choice.LatestSchemaStatus[k] = v
		}
	}

	return choice, nil
}

// SelectVersion applies the user's version selection to the
// session. If the session already has a version set
// (mid-session switch), the confirmer is called to get
// explicit user confirmation. If an MCP client is provided
// and connected, a version compatibility check is performed.
func SelectVersion(
	choice *VersionChoice,
	selection SelectionType,
	sess *session.Session,
	mcpClient *mcp.Client,
	confirmer VersionSwitchConfirmer,
) (*SelectionResult, error) {
	var target *Release
	switch selection {
	case SelectionStable:
		target = choice.StableVersion
	case SelectionLatest:
		target = choice.LatestVersion
	}

	if target == nil {
		return nil, ErrNoVersionAvailable
	}

	result := &SelectionResult{
		SelectedTag: target.Tag,
	}

	// Collect experimental schemas at selected version.
	result.ExperimentalSchemas = experimentalSchemas(
		target.SchemaStatusMap,
	)

	// Mid-session switch: require confirmation.
	if sess.SchemaVersion != "" &&
		sess.SchemaVersion != target.Tag {
		if confirmer == nil {
			return nil, fmt.Errorf(
				"mid-session version switch from %s "+
					"to %s requires confirmation but "+
					"no confirmer provided",
				sess.SchemaVersion,
				target.Tag,
			)
		}

		confirmed, err := confirmer(
			sess.SchemaVersion, target.Tag,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"confirmation failed: %w", err,
			)
		}
		if !confirmed {
			return nil, ErrVersionSwitchDeclined
		}
		result.PreviousVersion = sess.SchemaVersion
	}

	// Apply version to session.
	sess.SchemaVersion = target.Tag

	// MCP compatibility check if client is connected.
	if mcpClient != nil &&
		mcpClient.Status() == mcp.StatusConnected {
		compatResult, err := mcp.CheckCompatibility(
			context.Background(),
			mcp.DefaultVersionFetcher,
			mcpClient,
			target.Tag,
		)
		if err == nil && compatResult != nil &&
			compatResult.Status != mcp.CompatOK {
			result.CompatWarning =
				compatResult.Recommendation
		}
	}

	return result, nil
}

// experimentalSchemas returns a sorted list of schema names
// marked Experimental in the given status map.
func experimentalSchemas(
	statusMap map[string]SchemaStatus,
) []string {
	var names []string
	for name, status := range statusMap {
		if status == StatusExperimental {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
