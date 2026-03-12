// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hbraswelrh/pacman/internal/cli"
	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/mcp"
	"github.com/hbraswelrh/pacman/internal/session"
)

// mockPrompter returns predetermined choices.
type mockPrompter struct {
	choices []int
	idx     int
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
		// Create fake binary on make build.
		if name == "make" && len(args) > 0 &&
			args[0] == "build" && dir != "" {
			binaryPath := filepath.Join(
				dir, consts.MCPBinaryName,
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
// returns a connected session.
func TestSetup_MCPAlreadyDetected(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")
	var buf bytes.Buffer

	cfg := &cli.SetupConfig{
		Prompter:      &mockPrompter{},
		BinaryLookup:  mockBinaryFound("/usr/local/bin/gemara-mcp"),
		PodmanChecker: mockPodmanNotRunning(),
		ConfigPath:    configPath,
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

	// Verify opencode.json was configured.
	config, err := mcp.ReadOpenCodeConfig(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	entry, ok := config.MCP[consts.MCPServerName]
	if !ok {
		t.Fatal("expected gemara-mcp entry in config")
	}
	if len(entry.Command) != 1 ||
		entry.Command[0] != "/usr/local/bin/gemara-mcp" {
		t.Fatalf("unexpected command: %v", entry.Command)
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
			choices: []int{0, 0}, // Build from source, SSH
		},
		BinaryLookup:  mockBinaryNotFound(),
		PodmanChecker: mockPodmanNotRunning(),
		Installer:     mockInstaller(t),
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
	if result.Session.GetMCPStatus() != session.MCPConnected {
		t.Fatal("expected MCPConnected")
	}

	output := buf.String()
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

// TestSetup_PodmanInstall verifies the Podman installation
// flow.
func TestSetup_PodmanInstall(t *testing.T) {
	var buf bytes.Buffer
	var commands []string

	runner := func(
		_ context.Context,
		_ string,
		name string,
		args ...string,
	) ([]byte, error) {
		cmd := fmt.Sprintf("%s %v", name, args)
		commands = append(commands, cmd)
		return nil, nil
	}

	cfg := &cli.SetupConfig{
		Prompter: &mockPrompter{
			choices: []int{1}, // Podman
		},
		BinaryLookup:  mockBinaryNotFound(),
		PodmanChecker: mockPodmanNotRunning(),
		Installer:     mcp.NewInstaller(nil, runner),
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
	if !strings.Contains(output, "Podman container running") {
		t.Fatalf(
			"expected podman running message, got: %s",
			output,
		)
	}
}

// TestSetup_ExplainsMCPTools verifies that the setup prompt
// explains all three MCP tools.
func TestSetup_ExplainsMCPTools(t *testing.T) {
	var buf bytes.Buffer

	cfg := &cli.SetupConfig{
		Prompter: &mockPrompter{
			choices: []int{2}, // Skip
		},
		BinaryLookup:  mockBinaryNotFound(),
		PodmanChecker: mockPodmanNotRunning(),
	}

	_, err := cli.RunSetup(
		context.Background(), cfg, &buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	tools := []string{
		consts.ToolGetLexicon,
		consts.ToolValidateArtifact,
		consts.ToolGetSchemaDocs,
	}
	for _, tool := range tools {
		if !strings.Contains(output, tool) {
			t.Fatalf(
				"expected tool %q in output, got: %s",
				tool,
				output,
			)
		}
	}
}
