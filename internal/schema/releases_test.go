// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/hbraswelrh/journey/internal/schema"
)

// mockHTTPClient implements schema.HTTPClient for testing.
type mockHTTPClient struct {
	resp *http.Response
	err  error
}

func (m *mockHTTPClient) Do(
	_ *http.Request,
) (*http.Response, error) {
	return m.resp, m.err
}

func jsonResponse(
	status int,
	body string,
) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func TestFetchReleases_Success(t *testing.T) {
	client := &mockHTTPClient{
		resp: jsonResponse(http.StatusOK, `[
			{
				"tag_name": "v0.20.0",
				"target_commitish": "abc123",
				"published_at": "2026-03-01T00:00:00Z"
			},
			{
				"tag_name": "v0.19.0",
				"target_commitish": "def456",
				"published_at": "2026-02-01T00:00:00Z"
			}
		]`),
	}

	releases, err := schema.FetchReleases(
		context.Background(), client,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(releases) != 2 {
		t.Fatalf(
			"expected 2 releases, got %d",
			len(releases),
		)
	}
	// Verify sorted newest first.
	if releases[0].Tag != "v0.20.0" {
		t.Fatalf(
			"expected v0.20.0 first, got %s",
			releases[0].Tag,
		)
	}
}

func TestFetchReleases_EmptyList(t *testing.T) {
	client := &mockHTTPClient{
		resp: jsonResponse(http.StatusOK, `[]`),
	}

	releases, err := schema.FetchReleases(
		context.Background(), client,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(releases) != 0 {
		t.Fatalf(
			"expected 0 releases, got %d",
			len(releases),
		)
	}
}

func TestFetchReleases_APIError(t *testing.T) {
	client := &mockHTTPClient{
		err: errors.New("network unreachable"),
	}

	_, err := schema.FetchReleases(
		context.Background(), client,
	)
	if err == nil {
		t.Fatal("expected error for unreachable API")
	}
}

func TestFetchReleases_Non200(t *testing.T) {
	client := &mockHTTPClient{
		resp: jsonResponse(
			http.StatusForbidden,
			`{"message": "rate limited"}`,
		),
	}

	_, err := schema.FetchReleases(
		context.Background(), client,
	)
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestDetermineStableVersion(t *testing.T) {
	releases := []schema.Release{
		{
			Tag:  "v0.20.0",
			Date: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base":           schema.StatusExperimental,
				"metadata":       schema.StatusStable,
				"mapping_inline": schema.StatusStable,
			},
		},
		{
			Tag:  "v0.19.0",
			Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base":           schema.StatusStable,
				"metadata":       schema.StatusStable,
				"mapping_inline": schema.StatusStable,
			},
		},
		{
			Tag:  "v0.18.0",
			Date: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base":           schema.StatusStable,
				"metadata":       schema.StatusStable,
				"mapping_inline": schema.StatusStable,
			},
		},
	}

	stable := schema.DetermineStableVersion(releases)
	if stable == nil {
		t.Fatal("expected a stable version")
	}
	// v0.20.0 has Experimental base, so stable should be
	// v0.19.0.
	if stable.Tag != "v0.19.0" {
		t.Fatalf(
			"expected v0.19.0 as stable, got %s",
			stable.Tag,
		)
	}
}

func TestDetermineLatestVersion(t *testing.T) {
	releases := []schema.Release{
		{
			Tag:  "v0.20.0",
			Date: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Tag:  "v0.19.0",
			Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	latest := schema.DetermineLatestVersion(releases)
	if latest == nil {
		t.Fatal("expected a latest version")
	}
	if latest.Tag != "v0.20.0" {
		t.Fatalf(
			"expected v0.20.0 as latest, got %s",
			latest.Tag,
		)
	}
}

func TestDetermineLatestVersion_Empty(t *testing.T) {
	latest := schema.DetermineLatestVersion(nil)
	if latest != nil {
		t.Fatal("expected nil for empty releases")
	}
}

func TestDetermineStableVersion_NoneStable(t *testing.T) {
	releases := []schema.Release{
		{
			Tag: "v0.20.0",
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base":           schema.StatusExperimental,
				"metadata":       schema.StatusExperimental,
				"mapping_inline": schema.StatusExperimental,
			},
		},
	}

	stable := schema.DetermineStableVersion(releases)
	if stable != nil {
		t.Fatal("expected nil when no version is stable")
	}
}
