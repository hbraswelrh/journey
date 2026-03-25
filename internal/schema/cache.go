// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// CachedReleases holds cached release data with a timestamp.
type CachedReleases struct {
	// Releases is the list of cached releases.
	Releases []Release `json:"releases"`
	// LastFetched is the timestamp of the last successful
	// upstream fetch.
	LastFetched time.Time `json:"last_fetched"`
	// FromCache is true if this data was loaded from cache
	// rather than fetched live.
	FromCache bool `json:"-"`
}

// WriteCache serializes release data to a JSON file.
func WriteCache(
	path string,
	releases []Release,
	fetchTime time.Time,
) error {
	cached := CachedReleases{
		Releases:    releases,
		LastFetched: fetchTime,
	}
	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

// ReadCache loads cached release data from a JSON file.
func ReadCache(path string) (*CachedReleases, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read cache: %w", err)
	}

	var cached CachedReleases
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("parse cache: %w", err)
	}
	cached.FromCache = true
	return &cached, nil
}

// RefreshOrCache attempts to fetch releases from upstream. If
// successful, it updates the cache. If upstream is unreachable,
// it falls back to the local cache.
func RefreshOrCache(
	ctx context.Context,
	fetcher func(ctx context.Context) ([]Release, error),
	cachePath string,
) (*CachedReleases, error) {
	// Try upstream first.
	releases, err := fetcher(ctx)
	if err == nil && len(releases) > 0 {
		now := time.Now().UTC()
		// Update cache (best effort — don't fail if cache
		// write fails).
		_ = WriteCache(cachePath, releases, now)
		return &CachedReleases{
			Releases:    releases,
			LastFetched: now,
			FromCache:   false,
		}, nil
	}

	// Fall back to cache.
	cached, cacheErr := ReadCache(cachePath)
	if cacheErr != nil {
		if err != nil {
			return nil, fmt.Errorf(
				"upstream unreachable (%w) and no local "+
					"cache available (%v)",
				err,
				cacheErr,
			)
		}
		return nil, fmt.Errorf("read cache: %w", cacheErr)
	}

	return cached, nil
}
