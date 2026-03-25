// SPDX-License-Identifier: Apache-2.0

// Package schema handles Gemara schema version fetching,
// caching, and selection.
package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
)

// SchemaStatus represents whether a schema is Stable or
// Experimental at a given version.
type SchemaStatus string

const (
	StatusStable       SchemaStatus = "Stable"
	StatusExperimental SchemaStatus = "Experimental"
)

// Release represents a tagged release of the Gemara schema
// repository.
type Release struct {
	// Tag is the version tag (e.g., "v0.20.0").
	Tag string `json:"tag"`
	// CommitSHA is the commit digest for the release.
	CommitSHA string `json:"commit_sha"`
	// Date is the release publication date.
	Date time.Time `json:"date"`
	// SchemaStatusMap maps schema names to their status at
	// this version (Stable or Experimental).
	SchemaStatusMap map[string]SchemaStatus `json:"schema_status"`
}

// GitHubReleaseResponse represents a subset of the GitHub API
// release list response.
type GitHubReleaseResponse struct {
	TagName     string    `json:"tag_name"`
	TargetSHA   string    `json:"target_commitish"`
	PublishedAt time.Time `json:"published_at"`
}

// ReleaseFetcher abstracts release fetching for testing.
type ReleaseFetcher func(
	ctx context.Context,
) ([]Release, error)

// HTTPClient abstracts HTTP requests for testing.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// FetchReleases queries the GitHub API for releases of the
// Gemara schema repository and returns a parsed list sorted
// by date (newest first).
func FetchReleases(
	ctx context.Context,
	client HTTPClient,
) ([]Release, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet,
		consts.GemaraReleasesAPI,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"GitHub API returned %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var ghReleases []GitHubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(
		&ghReleases,
	); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	releases := make([]Release, 0, len(ghReleases))
	for _, ghr := range ghReleases {
		releases = append(releases, Release{
			Tag:             ghr.TagName,
			CommitSHA:       ghr.TargetSHA,
			Date:            ghr.PublishedAt,
			SchemaStatusMap: make(map[string]SchemaStatus),
		})
	}

	// Sort by date, newest first.
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Date.After(releases[j].Date)
	})

	return releases, nil
}

// DetermineStableVersion identifies the most recent release
// where all core schemas are marked Stable.
func DetermineStableVersion(
	releases []Release,
) *Release {
	for i := range releases {
		if isCoreStable(&releases[i]) {
			return &releases[i]
		}
	}
	return nil
}

// DetermineLatestVersion returns the most recent release.
func DetermineLatestVersion(
	releases []Release,
) *Release {
	if len(releases) == 0 {
		return nil
	}
	return &releases[0]
}

// isCoreStable checks if all core schemas are marked Stable
// at the given release.
func isCoreStable(r *Release) bool {
	for _, name := range consts.CoreStableSchemas {
		status, ok := r.SchemaStatusMap[name]
		if !ok || status != StatusStable {
			return false
		}
	}
	return true
}
