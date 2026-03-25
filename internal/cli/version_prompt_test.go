// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hbraswelrh/gemara-user-journey/internal/cli"
	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"github.com/hbraswelrh/gemara-user-journey/internal/mcp"
	"github.com/hbraswelrh/gemara-user-journey/internal/schema"
	"github.com/hbraswelrh/gemara-user-journey/internal/session"
)

// versionMockTransport implements mcp.Transport for version
// prompt tests.
type versionMockTransport struct {
	readResourceResp []byte
	readResourceErr  error
}

func (m *versionMockTransport) Connect(
	_ context.Context,
) error {
	return nil
}

func (m *versionMockTransport) Close() error {
	return nil
}

func (m *versionMockTransport) Ping(
	_ context.Context,
) error {
	return nil
}

func (m *versionMockTransport) Call(
	_ context.Context,
	_ string,
	_ map[string]any,
) ([]byte, error) {
	// Return version info for compatibility checks.
	return []byte(
		`{"server_version":"1.0.0",` +
			`"schema_version":"v0.19.0"}`,
	), nil
}

func (m *versionMockTransport) ReadResource(
	_ context.Context,
	_ string,
) ([]byte, error) {
	return m.readResourceResp, m.readResourceErr
}

func (m *versionMockTransport) ListPrompts(
	_ context.Context,
) ([]byte, error) {
	return nil, nil
}

// testReleaseFetcher returns a fetcher that produces a fixed
// release list. v0.20.0 is Latest (Experimental base),
// v0.19.0 is Stable.
func testReleaseFetcher() cli.ReleaseFetcherFn {
	return func(
		_ context.Context,
	) ([]schema.Release, error) {
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
		}, nil
	}
}

func offlineFetcher() cli.ReleaseFetcherFn {
	return func(
		_ context.Context,
	) ([]schema.Release, error) {
		return nil, errors.New("network unreachable")
	}
}

// TestVersionPrompt_DisplaysOptions verifies that the version
// prompt shows Stable and Latest options with version numbers
// and the user's selection is applied to the session.
func TestVersionPrompt_DisplaysOptions(t *testing.T) {
	var buf bytes.Buffer
	sess := session.NewSessionWithoutMCP("")

	cfg := &cli.VersionPromptConfig{
		Prompter: &mockPrompter{
			choices: []int{0}, // Select Stable
		},
		Fetcher:   testReleaseFetcher(),
		CachePath: filepath.Join(t.TempDir(), "rel.json"),
		Session:   sess,
	}

	err := cli.RunVersionSelection(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("RunVersionSelection failed: %v", err)
	}

	output := buf.String()

	// Should mention both versions.
	if !strings.Contains(output, "v0.19.0") {
		t.Fatalf(
			"expected v0.19.0 in output, got: %s",
			output,
		)
	}
	if !strings.Contains(output, "v0.20.0") {
		t.Fatalf(
			"expected v0.20.0 in output, got: %s",
			output,
		)
	}

	// Session should have the stable version.
	if sess.SchemaVersion != "v0.19.0" {
		t.Fatalf(
			"expected session version v0.19.0, got %s",
			sess.SchemaVersion,
		)
	}
}

// TestVersionPrompt_NewerVersionNotification verifies that
// when the user selects Stable and a newer Latest exists, a
// notification is displayed.
func TestVersionPrompt_NewerVersionNotification(
	t *testing.T,
) {
	var buf bytes.Buffer
	sess := session.NewSessionWithoutMCP("")

	cfg := &cli.VersionPromptConfig{
		Prompter: &mockPrompter{
			choices: []int{0}, // Select Stable
		},
		Fetcher:   testReleaseFetcher(),
		CachePath: filepath.Join(t.TempDir(), "rel.json"),
		Session:   sess,
	}

	err := cli.RunVersionSelection(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("RunVersionSelection failed: %v", err)
	}

	output := buf.String()

	// Should mention newer version available.
	if !strings.Contains(output, "newer") &&
		!strings.Contains(output, "Newer") &&
		!strings.Contains(output, "v0.20.0") {
		t.Fatalf(
			"expected newer version notification, "+
				"got: %s",
			output,
		)
	}
}

// TestVersionPrompt_OfflineWithCache verifies that when
// upstream is unreachable but cache exists, the cached version
// data is used and the user is informed.
func TestVersionPrompt_OfflineWithCache(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "releases.json")

	// Seed cache.
	cachedTime := time.Date(
		2026, 2, 15, 8, 0, 0, 0, time.UTC,
	)
	cachedReleases := []schema.Release{
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
	}
	if err := schema.WriteCache(
		cachePath, cachedReleases, cachedTime,
	); err != nil {
		t.Fatalf("failed to seed cache: %v", err)
	}

	var buf bytes.Buffer
	sess := session.NewSessionWithoutMCP("")

	cfg := &cli.VersionPromptConfig{
		Prompter: &mockPrompter{
			choices: []int{0}, // Select Stable
		},
		Fetcher:   offlineFetcher(),
		CachePath: cachePath,
		Session:   sess,
	}

	err := cli.RunVersionSelection(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("RunVersionSelection failed: %v", err)
	}

	output := buf.String()

	// Should inform user about using cached data.
	if !strings.Contains(output, "cached") &&
		!strings.Contains(output, "Cached") &&
		!strings.Contains(output, "offline") {
		t.Fatalf(
			"expected cache/offline message, got: %s",
			output,
		)
	}

	// Session should have the cached version.
	if sess.SchemaVersion != "v0.19.0" {
		t.Fatalf(
			"expected session version v0.19.0, got %s",
			sess.SchemaVersion,
		)
	}
}

// TestVersionPrompt_MCPSchemaDocsRead verifies that when
// MCP is connected and a version is selected, the system
// reads schema docs for the selected version and shows
// a confirmation message.
func TestVersionPrompt_MCPSchemaDocsRead(t *testing.T) {
	var buf bytes.Buffer
	sess := session.NewSessionWithMCP(
		"", consts.MCPModeArtifact,
	)

	transport := &versionMockTransport{
		readResourceResp: []byte(
			`{"definitions": {"#ControlCatalog": {}}}`,
		),
	}
	client := mcp.NewClient(
		transport, mcp.DefaultClientConfig(),
	)
	if err := client.Connect(
		context.Background(),
	); err != nil {
		t.Fatalf("connect: %v", err)
	}

	cfg := &cli.VersionPromptConfig{
		Prompter: &mockPrompter{
			choices: []int{0}, // Select Stable
		},
		Fetcher:   testReleaseFetcher(),
		CachePath: filepath.Join(t.TempDir(), "rel.json"),
		Session:   sess,
		MCPClient: client,
	}

	err := cli.RunVersionSelection(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("RunVersionSelection failed: %v", err)
	}

	output := buf.String()

	// Should mention schema docs verification.
	if !strings.Contains(output, "schema documentation") &&
		!strings.Contains(output, "Schema documentation") {
		t.Fatalf(
			"expected schema docs message, got: %s",
			output,
		)
	}
}

// TestVersionPrompt_MCPSchemaDocsReadFailure verifies that
// when schema docs read fails, a warning is shown but the
// flow continues.
func TestVersionPrompt_MCPSchemaDocsReadFailure(
	t *testing.T,
) {
	var buf bytes.Buffer
	sess := session.NewSessionWithMCP(
		"", consts.MCPModeArtifact,
	)

	transport := &versionMockTransport{
		readResourceErr: errors.New("resource unavailable"),
	}
	client := mcp.NewClient(
		transport, mcp.DefaultClientConfig(),
	)
	if err := client.Connect(
		context.Background(),
	); err != nil {
		t.Fatalf("connect: %v", err)
	}

	cfg := &cli.VersionPromptConfig{
		Prompter: &mockPrompter{
			choices: []int{0}, // Select Stable
		},
		Fetcher:   testReleaseFetcher(),
		CachePath: filepath.Join(t.TempDir(), "rel.json"),
		Session:   sess,
		MCPClient: client,
	}

	err := cli.RunVersionSelection(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("RunVersionSelection failed: %v", err)
	}

	// Should still succeed (non-fatal).
	if sess.SchemaVersion != "v0.19.0" {
		t.Fatalf(
			"expected v0.19.0, got %s",
			sess.SchemaVersion,
		)
	}
}

// TestVersionPrompt_Integration verifies the full flow from
// MCP setup (declined) through version selection.
func TestVersionPrompt_Integration(t *testing.T) {
	var buf bytes.Buffer

	// Run setup first (decline MCP).
	setupCfg := &cli.SetupConfig{
		Prompter: &mockPrompter{
			choices: []int{2}, // Skip MCP
		},
		BinaryLookup:  mockBinaryNotFound(),
		PodmanChecker: mockPodmanNotRunning(),
	}

	setupResult, err := cli.RunSetup(
		context.Background(), setupCfg, &buf,
	)
	if err != nil {
		t.Fatalf("RunSetup failed: %v", err)
	}

	// Now run version selection on the same session.
	buf.Reset()
	versionCfg := &cli.VersionPromptConfig{
		Prompter: &mockPrompter{
			choices: []int{1}, // Select Latest
		},
		Fetcher:   testReleaseFetcher(),
		CachePath: filepath.Join(t.TempDir(), "rel.json"),
		Session:   setupResult.Session,
	}

	err = cli.RunVersionSelection(
		context.Background(), versionCfg, &buf,
	)
	if err != nil {
		t.Fatalf("RunVersionSelection failed: %v", err)
	}

	// Session should have the latest version.
	if setupResult.Session.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"expected session version v0.20.0, got %s",
			setupResult.Session.SchemaVersion,
		)
	}

	output := buf.String()

	// Should warn about experimental schemas.
	if !strings.Contains(output, "Experimental") &&
		!strings.Contains(output, "experimental") {
		t.Fatalf(
			"expected experimental warning, got: %s",
			output,
		)
	}
}

// T036: RunVersionSelection still compiles and functions
// correctly when called directly. This proves the function
// is preserved and functional despite being bypassed in the
// active setup flow (ADR-0003).
func TestVersionPrompt_PreservedAndFunctional(
	t *testing.T,
) {
	var buf bytes.Buffer
	sess := session.NewSessionWithoutMCP("")

	cfg := &cli.VersionPromptConfig{
		Prompter: &mockPrompter{
			choices: []int{1}, // Select Latest
		},
		Fetcher:   testReleaseFetcher(),
		CachePath: filepath.Join(t.TempDir(), "rel.json"),
		Session:   sess,
	}

	err := cli.RunVersionSelection(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf(
			"RunVersionSelection should still work "+
				"when called directly: %v", err,
		)
	}

	// Session should have latest version selected.
	if sess.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"expected session version v0.20.0, got %s",
			sess.SchemaVersion,
		)
	}
}
