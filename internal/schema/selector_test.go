// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hbraswelrh/pacman/internal/mcp"
	"github.com/hbraswelrh/pacman/internal/schema"
	"github.com/hbraswelrh/pacman/internal/session"
)

// mockTransport implements mcp.Transport for selector tests.
type mockTransport struct {
	connectErr error
	pingErr    error
	callResp   []byte
	callErr    error
}

func (m *mockTransport) Connect(
	_ context.Context,
) error {
	return m.connectErr
}

func (m *mockTransport) Close() error {
	return nil
}

func (m *mockTransport) Ping(
	_ context.Context,
) error {
	return m.pingErr
}

func (m *mockTransport) Call(
	_ context.Context,
	_ string,
	_ map[string]any,
) ([]byte, error) {
	return m.callResp, m.callErr
}

func (m *mockTransport) ReadResource(
	_ context.Context,
	_ string,
) ([]byte, error) {
	return nil, nil
}

func (m *mockTransport) ListPrompts(
	_ context.Context,
) ([]byte, error) {
	return nil, nil
}

// testReleases returns a standard release list for selector
// tests. v0.20.0 has Experimental base (Latest only), v0.19.0
// is fully Stable, v0.18.0 is fully Stable but older.
func testReleases() []schema.Release {
	return []schema.Release{
		{
			Tag:       "v0.20.0",
			CommitSHA: "aaa111",
			Date: time.Date(
				2026, 3, 1, 0, 0, 0, 0, time.UTC,
			),
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base":           schema.StatusExperimental,
				"metadata":       schema.StatusStable,
				"mapping_inline": schema.StatusStable,
			},
		},
		{
			Tag:       "v0.19.0",
			CommitSHA: "bbb222",
			Date: time.Date(
				2026, 2, 1, 0, 0, 0, 0, time.UTC,
			),
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base":           schema.StatusStable,
				"metadata":       schema.StatusStable,
				"mapping_inline": schema.StatusStable,
			},
		},
		{
			Tag:       "v0.18.0",
			CommitSHA: "ccc333",
			Date: time.Date(
				2026, 1, 1, 0, 0, 0, 0, time.UTC,
			),
			SchemaStatusMap: map[string]schema.SchemaStatus{
				"base":           schema.StatusStable,
				"metadata":       schema.StatusStable,
				"mapping_inline": schema.StatusStable,
			},
		},
	}
}

func TestDetermineVersions_StableVersion(t *testing.T) {
	releases := testReleases()

	choice, err := schema.DetermineVersions(releases)
	if err != nil {
		t.Fatalf("DetermineVersions failed: %v", err)
	}

	if choice.StableVersion == nil {
		t.Fatal("expected a StableVersion")
	}
	if choice.StableVersion.Tag != "v0.19.0" {
		t.Fatalf(
			"expected stable v0.19.0, got %s",
			choice.StableVersion.Tag,
		)
	}

	// Verify status map is populated.
	status, ok := choice.StableSchemaStatus["base"]
	if !ok {
		t.Fatal("expected base in StableSchemaStatus")
	}
	if status != schema.StatusStable {
		t.Fatalf(
			"expected base Stable, got %s",
			status,
		)
	}
}

func TestDetermineVersions_LatestVersion(t *testing.T) {
	releases := testReleases()

	choice, err := schema.DetermineVersions(releases)
	if err != nil {
		t.Fatalf("DetermineVersions failed: %v", err)
	}

	if choice.LatestVersion == nil {
		t.Fatal("expected a LatestVersion")
	}
	if choice.LatestVersion.Tag != "v0.20.0" {
		t.Fatalf(
			"expected latest v0.20.0, got %s",
			choice.LatestVersion.Tag,
		)
	}

	// v0.20.0 has Experimental base.
	status, ok := choice.LatestSchemaStatus["base"]
	if !ok {
		t.Fatal("expected base in LatestSchemaStatus")
	}
	if status != schema.StatusExperimental {
		t.Fatalf(
			"expected base Experimental, got %s",
			status,
		)
	}
}

func TestSelectVersion_Stable(t *testing.T) {
	releases := testReleases()
	choice, err := schema.DetermineVersions(releases)
	if err != nil {
		t.Fatalf("DetermineVersions failed: %v", err)
	}

	sess := session.NewSessionWithoutMCP("")

	result, err := schema.SelectVersion(
		choice,
		schema.SelectionStable,
		sess,
		nil, // no MCP client
		nil, // no confirmer needed for fresh session
	)
	if err != nil {
		t.Fatalf("SelectVersion failed: %v", err)
	}

	if sess.SchemaVersion != "v0.19.0" {
		t.Fatalf(
			"expected session version v0.19.0, got %s",
			sess.SchemaVersion,
		)
	}
	if result.SelectedTag != "v0.19.0" {
		t.Fatalf(
			"expected result tag v0.19.0, got %s",
			result.SelectedTag,
		)
	}
	if len(result.ExperimentalSchemas) != 0 {
		t.Fatalf(
			"expected no experimental schemas for "+
				"stable, got %v",
			result.ExperimentalSchemas,
		)
	}
	if result.CompatWarning != "" {
		t.Fatalf(
			"expected no compat warning, got %s",
			result.CompatWarning,
		)
	}
}

func TestSelectVersion_Latest_ExperimentalWarning(
	t *testing.T,
) {
	releases := testReleases()
	choice, err := schema.DetermineVersions(releases)
	if err != nil {
		t.Fatalf("DetermineVersions failed: %v", err)
	}

	sess := session.NewSessionWithoutMCP("")

	result, err := schema.SelectVersion(
		choice,
		schema.SelectionLatest,
		sess,
		nil, // no MCP client
		nil, // no confirmer needed for fresh session
	)
	if err != nil {
		t.Fatalf("SelectVersion failed: %v", err)
	}

	if sess.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"expected session version v0.20.0, got %s",
			sess.SchemaVersion,
		)
	}

	// Should list experimental schemas.
	if len(result.ExperimentalSchemas) == 0 {
		t.Fatal(
			"expected experimental schemas to be listed",
		)
	}
	found := false
	for _, name := range result.ExperimentalSchemas {
		if name == "base" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf(
			"expected 'base' in experimental schemas, "+
				"got %v",
			result.ExperimentalSchemas,
		)
	}
}

func TestSelectVersion_Latest_MCPCompatCheck(
	t *testing.T,
) {
	releases := testReleases()
	choice, err := schema.DetermineVersions(releases)
	if err != nil {
		t.Fatalf("DetermineVersions failed: %v", err)
	}

	sess := session.NewSessionWithMCP("", "")

	// Create a mock MCP client that returns a mismatch.
	transport := &mockTransport{
		callResp: []byte(
			`{"server_version":"1.0.0",` +
				`"schema_version":"v0.19.0"}`,
		),
	}
	client := mcp.NewClient(transport, mcp.DefaultClientConfig())
	// Connect the client so callTool works.
	if err := client.Connect(
		t.Context(),
	); err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	result, err := schema.SelectVersion(
		choice,
		schema.SelectionLatest,
		sess,
		client,
		nil, // no confirmer needed for fresh session
	)
	if err != nil {
		t.Fatalf("SelectVersion failed: %v", err)
	}

	if sess.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"expected session version v0.20.0, got %s",
			sess.SchemaVersion,
		)
	}

	// Should have a compatibility warning.
	if result.CompatWarning == "" {
		t.Fatal(
			"expected compat warning for " +
				"MCP version mismatch",
		)
	}
}

func TestSelectVersion_MidSession_RequiresConfirmation(
	t *testing.T,
) {
	releases := testReleases()
	choice, err := schema.DetermineVersions(releases)
	if err != nil {
		t.Fatalf("DetermineVersions failed: %v", err)
	}

	// Session already has a version set (mid-session).
	sess := session.NewSessionWithoutMCP("v0.18.0")

	// User declines the switch.
	declineConfirmer := func(
		_, _ string,
	) (bool, error) {
		return false, nil
	}

	_, err = schema.SelectVersion(
		choice,
		schema.SelectionStable,
		sess,
		nil,
		declineConfirmer,
	)
	if err == nil {
		t.Fatal(
			"expected error when user declines " +
				"mid-session switch",
		)
	}

	// Session version should remain unchanged.
	if sess.SchemaVersion != "v0.18.0" {
		t.Fatalf(
			"expected session version unchanged "+
				"at v0.18.0, got %s",
			sess.SchemaVersion,
		)
	}

	// User confirms the switch.
	confirmConfirmer := func(
		_, _ string,
	) (bool, error) {
		return true, nil
	}

	result, err := schema.SelectVersion(
		choice,
		schema.SelectionStable,
		sess,
		nil,
		confirmConfirmer,
	)
	if err != nil {
		t.Fatalf(
			"SelectVersion failed after confirm: %v",
			err,
		)
	}

	if sess.SchemaVersion != "v0.19.0" {
		t.Fatalf(
			"expected session version v0.19.0 "+
				"after confirm, got %s",
			sess.SchemaVersion,
		)
	}
	if result.PreviousVersion != "v0.18.0" {
		t.Fatalf(
			"expected previous version v0.18.0, got %s",
			result.PreviousVersion,
		)
	}
}

// T006: AutoSelectLatest sets session to latest version.
func TestAutoSelectLatest_SetsLatest(t *testing.T) {
	releases := testReleases()

	fetcher := func(
		_ context.Context,
	) ([]schema.Release, error) {
		return releases, nil
	}

	sess := session.NewSessionWithoutMCP("")
	cachePath := t.TempDir() + "/releases.json"

	result, err := schema.AutoSelectLatest(
		t.Context(), fetcher, cachePath, sess,
	)
	if err != nil {
		t.Fatalf("AutoSelectLatest failed: %v", err)
	}

	if sess.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"expected session version v0.20.0, got %s",
			sess.SchemaVersion,
		)
	}
	if result.SelectedTag != "v0.20.0" {
		t.Fatalf(
			"expected result tag v0.20.0, got %s",
			result.SelectedTag,
		)
	}
}

// T006: AutoSelectLatest returns error for empty releases.
func TestAutoSelectLatest_EmptyReleases(t *testing.T) {
	fetcher := func(
		_ context.Context,
	) ([]schema.Release, error) {
		return nil, nil
	}

	sess := session.NewSessionWithoutMCP("")
	cachePath := t.TempDir() + "/releases.json"

	_, err := schema.AutoSelectLatest(
		t.Context(), fetcher, cachePath, sess,
	)
	if err == nil {
		t.Fatal(
			"expected error for empty releases",
		)
	}
}

// T006: AutoSelectLatest detects experimental schemas.
func TestAutoSelectLatest_ExperimentalSchemas(
	t *testing.T,
) {
	releases := testReleases()

	fetcher := func(
		_ context.Context,
	) ([]schema.Release, error) {
		return releases, nil
	}

	sess := session.NewSessionWithoutMCP("")
	cachePath := t.TempDir() + "/releases.json"

	result, err := schema.AutoSelectLatest(
		t.Context(), fetcher, cachePath, sess,
	)
	if err != nil {
		t.Fatalf("AutoSelectLatest failed: %v", err)
	}

	// v0.20.0 has Experimental base.
	if len(result.ExperimentalSchemas) == 0 {
		t.Fatal(
			"expected experimental schemas to be " +
				"detected",
		)
	}

	found := false
	for _, name := range result.ExperimentalSchemas {
		if name == "base" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf(
			"expected 'base' in experimental schemas, "+
				"got %v",
			result.ExperimentalSchemas,
		)
	}
}

// T006: AutoSelectLatest falls back to cache when fetcher
// fails.
func TestAutoSelectLatest_CacheFallback(t *testing.T) {
	releases := testReleases()

	// First, populate the cache.
	cachePath := t.TempDir() + "/releases.json"
	err := schema.WriteCache(
		cachePath, releases, time.Now(),
	)
	if err != nil {
		t.Fatalf("WriteCache failed: %v", err)
	}

	// Fetcher that always fails.
	failFetcher := func(
		_ context.Context,
	) ([]schema.Release, error) {
		return nil, fmt.Errorf("network unreachable")
	}

	sess := session.NewSessionWithoutMCP("")

	result, err := schema.AutoSelectLatest(
		t.Context(), failFetcher, cachePath, sess,
	)
	if err != nil {
		t.Fatalf(
			"expected cache fallback, got error: %v",
			err,
		)
	}

	if sess.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"expected session version v0.20.0 from "+
				"cache, got %s",
			sess.SchemaVersion,
		)
	}
	if result.SelectedTag != "v0.20.0" {
		t.Fatalf(
			"expected result tag v0.20.0, got %s",
			result.SelectedTag,
		)
	}
}
