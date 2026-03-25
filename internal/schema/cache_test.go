// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hbraswelrh/journey/internal/schema"
)

func TestWriteCache_CreatesValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "releases.json")

	fetchTime := time.Date(
		2026, 3, 10, 12, 0, 0, 0, time.UTC,
	)
	releases := []schema.Release{
		{
			Tag:       "v0.20.0",
			CommitSHA: "abc123",
			Date: time.Date(
				2026, 3, 1, 0, 0, 0, 0, time.UTC,
			),
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base":     schema.StatusStable,
				"metadata": schema.StatusStable,
			},
		},
	}

	err := schema.WriteCache(path, releases, fetchTime)
	if err != nil {
		t.Fatalf("WriteCache failed: %v", err)
	}

	// Verify the file exists and is valid JSON.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}

	var cached schema.CachedReleases
	if err := json.Unmarshal(data, &cached); err != nil {
		t.Fatalf("cache file is not valid JSON: %v", err)
	}

	if len(cached.Releases) != 1 {
		t.Fatalf(
			"expected 1 release, got %d",
			len(cached.Releases),
		)
	}
	if cached.Releases[0].Tag != "v0.20.0" {
		t.Fatalf(
			"expected tag v0.20.0, got %s",
			cached.Releases[0].Tag,
		)
	}
	if !cached.LastFetched.Equal(fetchTime) {
		t.Fatalf(
			"expected last_fetched %v, got %v",
			fetchTime,
			cached.LastFetched,
		)
	}
}

func TestReadCache_ReturnsCachedData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "releases.json")

	fetchTime := time.Date(
		2026, 3, 10, 12, 0, 0, 0, time.UTC,
	)
	releases := []schema.Release{
		{
			Tag:       "v0.19.0",
			CommitSHA: "def456",
			Date: time.Date(
				2026, 2, 1, 0, 0, 0, 0, time.UTC,
			),
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base": schema.StatusStable,
			},
		},
		{
			Tag:       "v0.18.0",
			CommitSHA: "ghi789",
			Date: time.Date(
				2026, 1, 1, 0, 0, 0, 0, time.UTC,
			),
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base": schema.StatusExperimental,
			},
		},
	}

	// Write cache first.
	if err := schema.WriteCache(
		path, releases, fetchTime,
	); err != nil {
		t.Fatalf("WriteCache setup failed: %v", err)
	}

	// Read it back.
	cached, err := schema.ReadCache(path)
	if err != nil {
		t.Fatalf("ReadCache failed: %v", err)
	}

	if !cached.FromCache {
		t.Fatal("expected FromCache to be true")
	}
	if len(cached.Releases) != 2 {
		t.Fatalf(
			"expected 2 releases, got %d",
			len(cached.Releases),
		)
	}
	if cached.Releases[0].Tag != "v0.19.0" {
		t.Fatalf(
			"expected first tag v0.19.0, got %s",
			cached.Releases[0].Tag,
		)
	}
	if !cached.LastFetched.Equal(fetchTime) {
		t.Fatalf(
			"expected last_fetched %v, got %v",
			fetchTime,
			cached.LastFetched,
		)
	}
}

func TestRefreshOrCache_UpstreamAvailable(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "releases.json")

	// Seed stale cache.
	staleTime := time.Date(
		2026, 1, 1, 0, 0, 0, 0, time.UTC,
	)
	staleReleases := []schema.Release{
		{Tag: "v0.18.0", CommitSHA: "old111"},
	}
	if err := schema.WriteCache(
		cachePath, staleReleases, staleTime,
	); err != nil {
		t.Fatalf("failed to seed stale cache: %v", err)
	}

	// Upstream returns fresh data.
	freshReleases := []schema.Release{
		{
			Tag:       "v0.20.0",
			CommitSHA: "fresh999",
			Date: time.Date(
				2026, 3, 1, 0, 0, 0, 0, time.UTC,
			),
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base": schema.StatusStable,
			},
		},
	}
	fetcher := func(
		_ context.Context,
	) ([]schema.Release, error) {
		return freshReleases, nil
	}

	result, err := schema.RefreshOrCache(
		context.Background(), fetcher, cachePath,
	)
	if err != nil {
		t.Fatalf("RefreshOrCache failed: %v", err)
	}

	if result.FromCache {
		t.Fatal(
			"expected FromCache=false when upstream " +
				"is available",
		)
	}
	if len(result.Releases) != 1 {
		t.Fatalf(
			"expected 1 release, got %d",
			len(result.Releases),
		)
	}
	if result.Releases[0].Tag != "v0.20.0" {
		t.Fatalf(
			"expected v0.20.0, got %s",
			result.Releases[0].Tag,
		)
	}

	// Verify the cache file was updated.
	cached, err := schema.ReadCache(cachePath)
	if err != nil {
		t.Fatalf("failed to read updated cache: %v", err)
	}
	if cached.Releases[0].Tag != "v0.20.0" {
		t.Fatalf(
			"cache not updated: expected v0.20.0, got %s",
			cached.Releases[0].Tag,
		)
	}
}

func TestRefreshOrCache_OfflineWithCache(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "releases.json")

	// Seed cache with known data.
	cachedTime := time.Date(
		2026, 2, 15, 8, 0, 0, 0, time.UTC,
	)
	cachedReleases := []schema.Release{
		{
			Tag:       "v0.19.0",
			CommitSHA: "cached222",
			Date: time.Date(
				2026, 2, 1, 0, 0, 0, 0, time.UTC,
			),
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base": schema.StatusStable,
			},
		},
	}
	if err := schema.WriteCache(
		cachePath, cachedReleases, cachedTime,
	); err != nil {
		t.Fatalf("failed to seed cache: %v", err)
	}

	// Upstream is unreachable.
	fetcher := func(
		_ context.Context,
	) ([]schema.Release, error) {
		return nil, errors.New("network unreachable")
	}

	result, err := schema.RefreshOrCache(
		context.Background(), fetcher, cachePath,
	)
	if err != nil {
		t.Fatalf(
			"expected fallback to cache, got error: %v",
			err,
		)
	}

	if !result.FromCache {
		t.Fatal(
			"expected FromCache=true when offline " +
				"with cache",
		)
	}
	if len(result.Releases) != 1 {
		t.Fatalf(
			"expected 1 cached release, got %d",
			len(result.Releases),
		)
	}
	if result.Releases[0].Tag != "v0.19.0" {
		t.Fatalf(
			"expected cached v0.19.0, got %s",
			result.Releases[0].Tag,
		)
	}
	if !result.LastFetched.Equal(cachedTime) {
		t.Fatalf(
			"expected cached timestamp %v, got %v",
			cachedTime,
			result.LastFetched,
		)
	}
}

func TestRefreshOrCache_OfflineNoCache(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "nonexistent.json")

	// Upstream is unreachable and no cache exists.
	fetcher := func(
		_ context.Context,
	) ([]schema.Release, error) {
		return nil, errors.New("network unreachable")
	}

	_, err := schema.RefreshOrCache(
		context.Background(), fetcher, cachePath,
	)
	if err == nil {
		t.Fatal(
			"expected error when offline with no cache",
		)
	}

	// Verify the error message mentions both problems.
	errMsg := err.Error()
	if !containsSubstr(errMsg, "upstream unreachable") {
		t.Fatalf(
			"error should mention upstream: %s",
			errMsg,
		)
	}
	if !containsSubstr(errMsg, "no local cache") {
		t.Fatalf(
			"error should mention missing cache: %s",
			errMsg,
		)
	}
}

// containsSubstr checks if s contains substr.
func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) &&
		searchSubstr(s, substr)
}

// searchSubstr is a simple substring search.
func searchSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
