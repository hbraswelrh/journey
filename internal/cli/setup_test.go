// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"context"
	"errors"
	"os"
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

// mockPrompter returns predetermined choices.
type mockPrompter struct {
	choices []int
	texts   []string
	idx     int
	textIdx int
}

func (m *mockPrompter) Ask(
	_ string,
	_ []string,
) (int, error) {
	if m.idx >= len(m.choices) {
		return 0, errors.New("no more choices")
	}
	choice := m.choices[m.idx]
	m.idx++
	return choice, nil
}

func (m *mockPrompter) AskText(
	_ string,
) (string, error) {
	if m.textIdx >= len(m.texts) {
		return "", errors.New("no more texts")
	}
	text := m.texts[m.textIdx]
	m.textIdx++
	return text, nil
}

func mockBinaryFound(
	path string,
) mcp.BinaryLookup {
	return func(_ string) (string, error) {
		return path, nil
	}
}

func mockBinaryNotFound() mcp.BinaryLookup {
	return func(_ string) (string, error) {
		return "", errors.New("not found")
	}
}

func mockPodmanNotRunning() mcp.PodmanChecker {
	return func(_ string) (bool, error) {
		return false, nil
	}
}

func mockInstaller(
	t *testing.T,
) *mcp.Installer {
	t.Helper()
	fetcher := func(
		_ context.Context,
		_ string,
	) (*mcp.ReleaseInfo, error) {
		return &mcp.ReleaseInfo{
			Tag:       "v0.5.0",
			CommitSHA: "abc123def456abc123def456",
		}, nil
	}
	runner := func(
		_ context.Context,
		dir string,
		name string,
		args ...string,
	) ([]byte, error) {
		// Create clone dir on git clone.
		if name == "git" && len(args) >= 2 &&
			args[0] == "clone" {
			cloneDir := args[len(args)-1]
			if err := os.MkdirAll(
				cloneDir, 0o755,
			); err != nil {
				return nil, err
			}
		}
		// Create fake binary at bin/<name> matching
		// the gemara-mcp project layout.
		if name == "make" && len(args) > 0 &&
			args[0] == "build" && dir != "" {
			binDir := filepath.Join(dir, "bin")
			if mkErr := os.MkdirAll(
				binDir, 0o755,
			); mkErr != nil {
				return nil, mkErr
			}
			binaryPath := filepath.Join(
				binDir, consts.MCPBinaryName,
			)
			if err := os.WriteFile(
				binaryPath, []byte("fake"), 0o755,
			); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
	return mcp.NewInstaller(fetcher, runner)
}

// TestSetup_MCPAlreadyDetected verifies that when the MCP
// server is already installed, setup skips the prompt and
// returns a connected session without reconfiguring.
func TestSetup_MCPAlreadyDetected(t *testing.T) {
	var buf bytes.Buffer

	cfg := &cli.SetupConfig{
		Prompter:      &mockPrompter{},
		BinaryLookup:  mockBinaryFound("/usr/local/bin/gemara-mcp"),
		PodmanChecker: mockPodmanNotRunning(),
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Session.GetMCPStatus() != session.MCPConnected {
		t.Fatal("expected MCPConnected")
	}
	if result.MCPInstalled {
		t.Fatal("expected MCPInstalled to be false")
	}
	if !strings.Contains(buf.String(), "detected") {
		t.Fatalf(
			"expected detection message, got: %s",
			buf.String(),
		)
	}
}

// TestSetup_UserDeclinesInstallation verifies that declining
// produces a fallback session with degraded capabilities.
func TestSetup_UserDeclinesInstallation(t *testing.T) {
	var buf bytes.Buffer

	cfg := &cli.SetupConfig{
		Prompter: &mockPrompter{
			choices: []int{2}, // Skip installation
		},
		BinaryLookup:  mockBinaryNotFound(),
		PodmanChecker: mockPodmanNotRunning(),
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Declined {
		t.Fatal("expected Declined to be true")
	}
	if result.Session.GetMCPStatus() != session.MCPNotInstalled {
		t.Fatal("expected MCPNotInstalled")
	}
	if !result.Session.IsFallback() {
		t.Fatal("expected fallback mode")
	}

	output := buf.String()
	if !strings.Contains(output, "skipped") {
		t.Fatalf(
			"expected skip message, got: %s",
			output,
		)
	}
	if !strings.Contains(output, "Degraded") {
		t.Fatalf(
			"expected degraded capabilities, got: %s",
			output,
		)
	}
}

// TestSetup_SourceBuildSSH verifies the full source build flow
// via SSH.
func TestSetup_SourceBuildSSH(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	// Override userHomeDir for test.
	origHome := cli.ExportUserHomeDir()
	cli.SetUserHomeDir(func() (string, error) {
		return dir, nil
	})
	defer cli.SetUserHomeDir(origHome)

	var buf bytes.Buffer
	cfg := &cli.SetupConfig{
		Prompter: &mockPrompter{
			// Build from source, artifact mode, confirm
			choices: []int{0, 0, 0},
		},
		BinaryLookup:  mockBinaryNotFound(),
		PodmanChecker: mockPodmanNotRunning(),
		Installer:     mockInstaller(t),
		SSHChecker: func(
			_ context.Context,
		) bool {
			return true // Simulate SSH keys configured
		},
		ConfigPath: configPath,
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.MCPInstalled {
		t.Fatal("expected MCPInstalled to be true")
	}
	if result.Session.GetMCPStatus() != session.MCPConnected {
		t.Fatal("expected MCPConnected")
	}

	output := buf.String()
	if !strings.Contains(output, "Using SSH") {
		t.Fatalf(
			"expected SSH detection message, got: %s",
			output,
		)
	}
	if !strings.Contains(output, "Build complete") {
		t.Fatalf(
			"expected build complete message, got: %s",
			output,
		)
	}
	if !strings.Contains(output, "configuration updated") {
		t.Fatalf(
			"expected config update message, got: %s",
			output,
		)
	}
}

// TestSetup_ExistingBinary verifies that users can provide
// a path to an existing gemara-mcp binary.
func TestSetup_ExistingBinary(t *testing.T) {
	var buf bytes.Buffer

	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	binaryPath := "/opt/gemara-mcp/bin/gemara-mcp"

	cfg := &cli.SetupConfig{
		Prompter: &mockPrompter{
			// "I already have it", artifact mode
			choices: []int{1, 0},
			texts:   []string{binaryPath},
		},
		BinaryLookup:  mockBinaryNotFound(),
		PodmanChecker: mockPodmanNotRunning(),
		ConfigPath:    configPath,
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.MCPInstalled {
		t.Fatal("expected MCPInstalled to be true")
	}

	output := buf.String()
	if !strings.Contains(output, binaryPath) {
		t.Fatalf(
			"expected binary path in output, got: %s",
			output,
		)
	}

	// Verify MCP config was written with the provided
	// path.
	config, err := mcp.ReadOpenCodeConfig(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	entry, ok := config.MCP[consts.MCPServerName]
	if !ok {
		t.Fatal("expected gemara-mcp entry in config")
	}
	gotPath := mcp.MCPBinaryPath(entry)
	if gotPath != binaryPath {
		t.Errorf(
			"command[0] = %q, want %q",
			gotPath, binaryPath,
		)
	}

	// Verify update guidance is shown.
	if !strings.Contains(output, "Keeping gemara-mcp") {
		t.Fatalf(
			"expected update guidance, got: %s",
			output,
		)
	}
}

// TestSetup_SourceBuild_ArtifactMode verifies that after
// a source build, the user is prompted to select server mode
// and artifact mode is correctly recorded.
func TestSetup_SourceBuild_ArtifactMode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	origHome := cli.ExportUserHomeDir()
	cli.SetUserHomeDir(func() (string, error) {
		return dir, nil
	})
	defer cli.SetUserHomeDir(origHome)

	var buf bytes.Buffer
	cfg := &cli.SetupConfig{
		Prompter: &mockPrompter{
			// Build from source, artifact mode, confirm
			choices: []int{0, 0, 0},
		},
		BinaryLookup:  mockBinaryNotFound(),
		PodmanChecker: mockPodmanNotRunning(),
		Installer:     mockInstaller(t),
		SSHChecker: func(
			_ context.Context,
		) bool {
			return true
		},
		ConfigPath: configPath,
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.MCPInstalled {
		t.Fatal("expected MCPInstalled")
	}
	if result.Session.GetServerMode() !=
		consts.MCPModeArtifact {
		t.Fatalf(
			"expected mode %q, got %q",
			consts.MCPModeArtifact,
			result.Session.GetServerMode(),
		)
	}
	if !result.Session.IsArtifactMode() {
		t.Fatal("expected IsArtifactMode true")
	}

	// Verify config has correct mode.
	config, err := mcp.ReadOpenCodeConfig(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	entry := config.MCP[consts.MCPServerName]
	mode := mcp.ParseMCPMode(entry)
	if mode != consts.MCPModeArtifact {
		t.Fatalf(
			"config mode = %q, want %q",
			mode, consts.MCPModeArtifact,
		)
	}

	output := buf.String()
	if !strings.Contains(output, "Server Mode") {
		t.Fatalf(
			"expected mode selection heading, got: %s",
			output,
		)
	}
}

// TestSetup_SourceBuild_AdvisoryMode verifies that selecting
// advisory mode records it correctly and informs the user
// that prompts are unavailable.
func TestSetup_SourceBuild_AdvisoryMode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	origHome := cli.ExportUserHomeDir()
	cli.SetUserHomeDir(func() (string, error) {
		return dir, nil
	})
	defer cli.SetUserHomeDir(origHome)

	var buf bytes.Buffer
	cfg := &cli.SetupConfig{
		Prompter: &mockPrompter{
			// Build from source, advisory mode, confirm
			choices: []int{0, 1, 0},
		},
		BinaryLookup:  mockBinaryNotFound(),
		PodmanChecker: mockPodmanNotRunning(),
		Installer:     mockInstaller(t),
		SSHChecker: func(
			_ context.Context,
		) bool {
			return true
		},
		ConfigPath: configPath,
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Session.GetServerMode() !=
		consts.MCPModeAdvisory {
		t.Fatalf(
			"expected mode %q, got %q",
			consts.MCPModeAdvisory,
			result.Session.GetServerMode(),
		)
	}
	if result.Session.IsArtifactMode() {
		t.Fatal("expected IsArtifactMode false")
	}
	if result.Session.HasPrompts() {
		t.Fatal(
			"expected HasPrompts false in advisory mode",
		)
	}

	// Verify config has advisory mode.
	config, err := mcp.ReadOpenCodeConfig(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	entry := config.MCP[consts.MCPServerName]
	mode := mcp.ParseMCPMode(entry)
	if mode != consts.MCPModeAdvisory {
		t.Fatalf(
			"config mode = %q, want %q",
			mode, consts.MCPModeAdvisory,
		)
	}

	output := buf.String()
	if !strings.Contains(output, "advisory") {
		t.Fatalf(
			"expected advisory mode in output, got: %s",
			output,
		)
	}
}

// TestSetup_ExistingBinary_EmptyPath declines when empty
// path is provided.
func TestSetup_ExistingBinary_EmptyPath(t *testing.T) {
	var buf bytes.Buffer

	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	cfg := &cli.SetupConfig{
		Prompter: &mockPrompter{
			// "I already have it", then empty path
			choices: []int{1},
			texts:   []string{""},
		},
		BinaryLookup:  mockBinaryNotFound(),
		PodmanChecker: mockPodmanNotRunning(),
		ConfigPath:    configPath,
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty path means user declined.
	if result.MCPInstalled {
		t.Fatal(
			"expected MCPInstalled false for empty path",
		)
	}
}

// TestDoctor_ShowsMCPCapabilities verifies that the doctor
// output includes the MCP capabilities table when all
// checks pass.
func TestDoctor_ShowsMCPCapabilities(t *testing.T) {
	t.Parallel()

	config := &mcp.OpenCodeConfig{
		MCP: map[string]mcp.OpenCodeMCPEntry{
			consts.MCPServerName: {
				Type: "local",
				Command: []string{
					"/usr/local/bin/gemara-mcp",
					"serve", "--mode", "artifact",
				},
			},
		},
	}

	cfg := &cli.DoctorConfig{
		LookupBinary: mockLookup(map[string]string{
			"opencode":           "/usr/local/bin/opencode",
			"go":                 "/usr/local/go/bin/go",
			"cue":                "/usr/local/bin/cue",
			consts.MCPBinaryName: "/usr/local/bin/gemara-mcp",
			"git":                "/usr/bin/git",
		}),
		ReadConfig:   mockReadConfig(config, nil),
		ConfigPath:   "/project/opencode.json",
		TutorialsDir: mockTutorialsDir(t),
	}

	var buf bytes.Buffer
	ok := cli.RunDoctor(cfg, &buf)
	if !ok {
		t.Fatal("expected all checks to pass")
	}

	output := buf.String()
	capabilities := []string{
		consts.ToolValidateArtifact,
		consts.ResourceLexicon,
		consts.ResourceSchemaDefinitions,
		consts.WizardThreatAssessment,
		consts.WizardControlCatalog,
	}
	for _, cap := range capabilities {
		if !strings.Contains(output, cap) {
			t.Fatalf(
				"expected %q in doctor output, "+
					"got: %s",
				cap, output,
			)
		}
	}

	// Should prompt to start OpenCode.
	if !strings.Contains(output, "opencode") {
		t.Fatalf(
			"expected OpenCode start instructions, "+
				"got: %s",
			output,
		)
	}
}

// setupTestReleaseFetcher returns a test fetcher with
// v0.20.0 (Latest, Experimental base) and v0.19.0 (Stable).
func setupTestReleaseFetcher() cli.ReleaseFetcherFn {
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

// T019: RunSetup with VersionFetcher auto-selects latest
// version without prompting.
func TestSetup_AutoSelectsLatestVersion(t *testing.T) {
	var buf bytes.Buffer

	cfg := &cli.SetupConfig{
		Prompter:         &mockPrompter{},
		BinaryLookup:     mockBinaryFound("/usr/local/bin/gemara-mcp"),
		PodmanChecker:    mockPodmanNotRunning(),
		VersionFetcher:   setupTestReleaseFetcher(),
		VersionCachePath: filepath.Join(t.TempDir(), "releases.json"),
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Session should have latest version auto-selected.
	if result.Session.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"expected session version v0.20.0, got %s",
			result.Session.SchemaVersion,
		)
	}

	output := buf.String()

	// Should show the selected version.
	if !strings.Contains(output, "v0.20.0") {
		t.Fatalf(
			"expected version v0.20.0 in output, "+
				"got: %s", output,
		)
	}

	// Should show "latest" indicator.
	if !strings.Contains(output, "latest") {
		t.Fatalf(
			"expected 'latest' in output, got: %s",
			output,
		)
	}
}

// T013: RunSetup with RolePrompter populates
// Recommendations on the returned session's profile.
func TestSetup_RoleDiscoveryPopulatesRecommendations(
	t *testing.T,
) {
	var buf bytes.Buffer

	tutDir := "../tutorials/testdata/valid"

	cfg := &cli.SetupConfig{
		Prompter:         &mockPrompter{},
		BinaryLookup:     mockBinaryFound("/usr/local/bin/gemara-mcp"),
		PodmanChecker:    mockPodmanNotRunning(),
		VersionFetcher:   setupTestReleaseFetcher(),
		VersionCachePath: filepath.Join(t.TempDir(), "releases.json"),
		RolePrompter: &mockPrompter{
			// Security Engineer, then activity text.
			choices: []int{0},
			texts: []string{
				"CI/CD pipeline management and " +
					"threat modeling",
			},
		},
		TutorialsDir: tutDir,
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Session should have role profile set.
	if result.Session.GetRoleName() == "" {
		t.Fatal("expected non-empty role name")
	}

	// Session should have RecommendedArtifacts > 0
	// because Security Engineer + CI/CD resolves Layer 2
	// which maps to ThreatCatalog and ControlCatalog.
	if result.Session.RecommendedArtifacts == 0 {
		t.Fatal(
			"expected RecommendedArtifacts > 0 " +
				"after role discovery with CI/CD " +
				"activities",
		)
	}

	// Verify resolved layers include Layer 2.
	hasL2 := false
	for _, l := range result.Session.ResolvedLayers {
		if l == consts.LayerThreatsControls {
			hasL2 = true
			break
		}
	}
	if !hasL2 {
		t.Errorf(
			"expected Layer 2 in resolved layers, "+
				"got: %v",
			result.Session.ResolvedLayers,
		)
	}

	// Verify version was also auto-selected.
	if result.Session.SchemaVersion != "v0.20.0" {
		t.Errorf(
			"expected schema version v0.20.0, got %s",
			result.Session.SchemaVersion,
		)
	}

	output := buf.String()

	// Verify artifact recommendations were rendered.
	if !strings.Contains(
		output, consts.ArtifactThreatCatalog,
	) {
		t.Errorf(
			"expected %q in output, got: %s",
			consts.ArtifactThreatCatalog, output,
		)
	}
}

// T047: End-to-end quickstart verification: launch setup,
// confirm no version prompt, confirm artifact recommendations
// display, and confirm handoff summary references are present
// in the role discovery output.
func TestSetup_QuickstartEndToEnd(t *testing.T) {
	var buf bytes.Buffer

	tutDir := "../tutorials/testdata/valid"

	cfg := &cli.SetupConfig{
		Prompter:         &mockPrompter{},
		BinaryLookup:     mockBinaryFound("/usr/local/bin/gemara-mcp"),
		PodmanChecker:    mockPodmanNotRunning(),
		VersionFetcher:   setupTestReleaseFetcher(),
		VersionCachePath: filepath.Join(t.TempDir(), "releases.json"),
		RolePrompter: &mockPrompter{
			// Security Engineer, then CI/CD activities.
			choices: []int{0},
			texts: []string{
				"CI/CD pipeline management and " +
					"threat modeling",
			},
		},
		TutorialsDir: tutDir,
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// 1. No version prompt — version auto-selected.
	if result.Session.SchemaVersion != "v0.20.0" {
		t.Fatalf(
			"expected auto-selected version v0.20.0, "+
				"got %s",
			result.Session.SchemaVersion,
		)
	}
	if !strings.Contains(output, "v0.20.0") {
		t.Fatalf(
			"expected version in output, got: %s",
			output,
		)
	}

	// 2. Artifact recommendations displayed.
	if !strings.Contains(
		output, consts.ArtifactThreatCatalog,
	) {
		t.Errorf(
			"expected %q in output, got: %s",
			consts.ArtifactThreatCatalog, output,
		)
	}
	if !strings.Contains(
		output, consts.ArtifactControlCatalog,
	) {
		t.Errorf(
			"expected %q in output, got: %s",
			consts.ArtifactControlCatalog, output,
		)
	}

	// 3. MCP wizard names displayed.
	if !strings.Contains(
		output, consts.WizardThreatAssessment,
	) {
		t.Errorf(
			"expected wizard %q in output, got: %s",
			consts.WizardThreatAssessment, output,
		)
	}
	if !strings.Contains(
		output, consts.WizardControlCatalog,
	) {
		t.Errorf(
			"expected wizard %q in output, got: %s",
			consts.WizardControlCatalog, output,
		)
	}

	// 4. Artifact descriptions displayed.
	threatDesc :=
		consts.ArtifactDescriptions[consts.ArtifactThreatCatalog]
	if !strings.Contains(output, threatDesc) {
		t.Errorf(
			"expected threat description in output, "+
				"got: %s",
			output,
		)
	}

	// 5. Session has role profile populated.
	if result.Session.GetRoleName() !=
		consts.RoleSecurityEngineer {
		t.Errorf(
			"expected role %s, got %s",
			consts.RoleSecurityEngineer,
			result.Session.GetRoleName(),
		)
	}

	// 6. Recommended artifacts count stored in session.
	if result.Session.RecommendedArtifacts == 0 {
		t.Error(
			"expected RecommendedArtifacts > 0",
		)
	}

	// 7. Learning path steps generated.
	if result.Session.LearningPathSteps == 0 {
		t.Error(
			"expected LearningPathSteps > 0",
		)
	}
}

// T020: RunSetup continues gracefully when AutoSelectLatest
// fails (no network, no cache).
func TestSetup_AutoSelectFailsGracefully(t *testing.T) {
	var buf bytes.Buffer

	failFetcher := func(
		_ context.Context,
	) ([]schema.Release, error) {
		return nil, errors.New("network unreachable")
	}

	cfg := &cli.SetupConfig{
		Prompter:         &mockPrompter{},
		BinaryLookup:     mockBinaryFound("/usr/local/bin/gemara-mcp"),
		PodmanChecker:    mockPodmanNotRunning(),
		VersionFetcher:   failFetcher,
		VersionCachePath: filepath.Join(t.TempDir(), "nonexistent.json"),
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf(
			"expected graceful degradation, got error: %v",
			err,
		)
	}

	// Session should have empty version (graceful fail).
	if result.Session.SchemaVersion != "" {
		t.Fatalf(
			"expected empty schema version, got %s",
			result.Session.SchemaVersion,
		)
	}

	output := buf.String()

	// Should contain a warning about the failure.
	if !strings.Contains(
		strings.ToLower(output), "warning",
	) && !strings.Contains(output, "could not") {
		t.Fatalf(
			"expected warning in output, got: %s",
			output,
		)
	}
}

// T048: End-to-end edge case — no network and no cache,
// verify graceful degradation with role discovery still
// functioning despite empty schema version.
func TestSetup_NoNetworkNoCacheWithRoleDiscovery(
	t *testing.T,
) {
	var buf bytes.Buffer

	tutDir := "../tutorials/testdata/valid"

	failFetcher := func(
		_ context.Context,
	) ([]schema.Release, error) {
		return nil, errors.New("network unreachable")
	}

	cfg := &cli.SetupConfig{
		Prompter:         &mockPrompter{},
		BinaryLookup:     mockBinaryFound("/usr/local/bin/gemara-mcp"),
		PodmanChecker:    mockPodmanNotRunning(),
		VersionFetcher:   failFetcher,
		VersionCachePath: filepath.Join(t.TempDir(), "nonexistent.json"),
		RolePrompter: &mockPrompter{
			// Security Engineer, then CI/CD activities.
			choices: []int{0},
			texts: []string{
				"CI/CD pipeline management and " +
					"threat modeling",
			},
		},
		TutorialsDir: tutDir,
	}

	result, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf(
			"expected graceful degradation, "+
				"got error: %v",
			err,
		)
	}

	output := buf.String()

	// 1. Version should be empty (network failed, no
	//    cache).
	if result.Session.SchemaVersion != "" {
		t.Fatalf(
			"expected empty schema version, got %s",
			result.Session.SchemaVersion,
		)
	}

	// 2. Warning message should appear about the version
	//    resolution failure.
	if !strings.Contains(
		strings.ToLower(output), "warning",
	) && !strings.Contains(output, "could not") {
		t.Errorf(
			"expected warning in output, got: %s",
			output,
		)
	}

	// 3. Role discovery should still complete
	//    successfully despite no schema version.
	if result.Session.GetRoleName() == "" {
		t.Error(
			"expected role discovery to complete " +
				"even without schema version",
		)
	}

	// 4. Artifact recommendations should still be
	//    populated — they depend on layers, not schema
	//    version.
	if result.Session.RecommendedArtifacts == 0 {
		t.Error(
			"expected artifact recommendations " +
				"despite schema version failure",
		)
	}

	// 5. Tutorials should load (though no version
	//    mismatch checks without a version).
	if result.Session.LearningPathSteps == 0 {
		t.Error(
			"expected learning path steps " +
				"despite schema version failure",
		)
	}
}
