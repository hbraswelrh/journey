// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hbraswelrh/gemara-user-journey/internal/cli"
	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"github.com/hbraswelrh/gemara-user-journey/internal/mcp"
)

// mockLookup returns a lookup function that resolves
// specific binaries.
func mockLookup(
	found map[string]string,
) func(string) (string, error) {
	return func(name string) (string, error) {
		if path, ok := found[name]; ok {
			return path, nil
		}
		return "", errors.New("not found")
	}
}

// mockReadConfig returns a config reader with the given
// config.
func mockReadConfig(
	config *mcp.OpenCodeConfig,
	err error,
) func(string) (*mcp.OpenCodeConfig, error) {
	return func(_ string) (*mcp.OpenCodeConfig, error) {
		return config, err
	}
}

func mockTutorialsDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "tutorials")
	if err := os.MkdirAll(
		filepath.Join(dir, "controls"), 0o755,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(
		filepath.Join(dir, "guidance"), 0o755,
	); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestDoctor_AllPass(t *testing.T) {
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
	if !strings.Contains(output, "All checks passed") {
		t.Fatalf(
			"expected 'All checks passed', got: %s",
			output,
		)
	}
}

func TestDoctor_OpenCodeMissing(t *testing.T) {
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
			"go":                 "/usr/local/go/bin/go",
			"cue":                "/usr/local/bin/cue",
			consts.MCPBinaryName: "/usr/local/bin/gemara-mcp",
			"git":                "/usr/bin/git",
			// opencode missing
		}),
		ReadConfig: mockReadConfig(config, nil),
		ConfigPath: "/project/opencode.json",
	}

	var buf bytes.Buffer
	ok := cli.RunDoctor(cfg, &buf)

	if ok {
		t.Fatal("expected failure when opencode missing")
	}

	output := buf.String()
	if !strings.Contains(output, "opencode not found") {
		t.Fatalf(
			"expected 'opencode not found', got: %s",
			output,
		)
	}
	if !strings.Contains(output, "brew install") {
		t.Fatalf("expected fix instructions, got: %s", output)
	}
}

func TestDoctor_NoMCPConfig(t *testing.T) {
	t.Parallel()

	config := &mcp.OpenCodeConfig{
		MCP: map[string]mcp.OpenCodeMCPEntry{},
	}

	cfg := &cli.DoctorConfig{
		LookupBinary: mockLookup(map[string]string{
			"opencode":           "/usr/local/bin/opencode",
			"go":                 "/usr/local/go/bin/go",
			"cue":                "/usr/local/bin/cue",
			consts.MCPBinaryName: "/usr/local/bin/gemara-mcp",
			"git":                "/usr/bin/git",
		}),
		ReadConfig: mockReadConfig(config, nil),
		ConfigPath: "/project/opencode.json",
	}

	var buf bytes.Buffer
	ok := cli.RunDoctor(cfg, &buf)

	if ok {
		t.Fatal("expected failure when MCP entry missing")
	}

	output := buf.String()
	if !strings.Contains(output, "no gemara-mcp entry") {
		t.Fatalf(
			"expected 'no gemara-mcp entry', got: %s",
			output,
		)
	}
}

func TestDoctor_MissingModeFlag(t *testing.T) {
	t.Parallel()

	config := &mcp.OpenCodeConfig{
		MCP: map[string]mcp.OpenCodeMCPEntry{
			consts.MCPServerName: {
				Type: "local",
				Command: []string{
					"/usr/local/bin/gemara-mcp",
					"serve",
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
		ReadConfig: mockReadConfig(config, nil),
		ConfigPath: "/project/opencode.json",
	}

	var buf bytes.Buffer
	_ = cli.RunDoctor(cfg, &buf)

	output := buf.String()
	if !strings.Contains(output, "no --mode flag") {
		t.Fatalf(
			"expected '--mode flag' warning, got: %s",
			output,
		)
	}
}

func TestDoctor_AdvisoryMode(t *testing.T) {
	t.Parallel()

	config := &mcp.OpenCodeConfig{
		MCP: map[string]mcp.OpenCodeMCPEntry{
			consts.MCPServerName: {
				Type: "local",
				Command: []string{
					"/usr/local/bin/gemara-mcp",
					"serve", "--mode", "advisory",
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
		t.Fatal("expected pass (advisory is valid)")
	}

	output := buf.String()
	if !strings.Contains(
		output, "wizard prompts disabled",
	) {
		t.Fatalf(
			"expected 'wizard prompts disabled', "+
				"got: %s",
			output,
		)
	}
}

func TestDoctor_CUEMissing_IsWarning(t *testing.T) {
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
			consts.MCPBinaryName: "/usr/local/bin/gemara-mcp",
			"git":                "/usr/bin/git",
			// cue missing
		}),
		ReadConfig:   mockReadConfig(config, nil),
		ConfigPath:   "/project/opencode.json",
		TutorialsDir: mockTutorialsDir(t),
	}

	var buf bytes.Buffer
	ok := cli.RunDoctor(cfg, &buf)

	// CUE missing is a warning, not a failure.
	if !ok {
		t.Fatal(
			"expected pass (CUE is only a warning)",
		)
	}

	output := buf.String()
	if !strings.Contains(output, "cue not found") {
		t.Fatalf(
			"expected 'cue not found' warning, got: %s",
			output,
		)
	}
}
