// SPDX-License-Identifier: Apache-2.0

package mcp_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/mcp"
)

func mockFetcher(
	tag, sha string,
) mcp.ReleaseFetcher {
	return func(
		_ context.Context,
		_ string,
	) (*mcp.ReleaseInfo, error) {
		return &mcp.ReleaseInfo{
			Tag:       tag,
			CommitSHA: sha,
		}, nil
	}
}

func mockRunner(
	commands *[]string,
	binaryName string,
) mcp.CommandRunner {
	return func(
		_ context.Context,
		dir string,
		name string,
		args ...string,
	) ([]byte, error) {
		cmd := fmt.Sprintf(
			"%s %s", name,
			fmt.Sprintf("%v", args),
		)
		*commands = append(*commands, cmd)

		// When git clone is called, create the target
		// directory so subsequent commands can operate in it.
		if name == "git" && len(args) >= 2 &&
			args[0] == "clone" {
			cloneDir := args[len(args)-1]
			if err := os.MkdirAll(
				cloneDir, 0o755,
			); err != nil {
				return nil, err
			}
		}

		// When make build is called, create a fake binary
		// in the working directory.
		if name == "make" && len(args) > 0 &&
			args[0] == "build" && dir != "" {
			binaryPath := filepath.Join(
				dir, binaryName,
			)
			if err := os.WriteFile(
				binaryPath, []byte("fake"), 0o755,
			); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
}

func TestInstaller_ResolveLatestRelease(t *testing.T) {
	fetcher := mockFetcher(
		"v0.5.0",
		"abc123def456abc123def456abc123def456abc1",
	)
	installer := mcp.NewInstaller(fetcher, nil)

	ctx := context.Background()
	info, err := installer.ResolveLatestRelease(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Tag != "v0.5.0" {
		t.Fatalf("expected tag v0.5.0, got %s", info.Tag)
	}
	if info.CommitSHA == "" {
		t.Fatal("expected non-empty CommitSHA")
	}
}

func TestInstaller_CloneAndBuild_SSH(t *testing.T) {
	destDir := t.TempDir()
	var commands []string
	sha := "abc123def456"

	fetcher := mockFetcher("v0.5.0", sha)
	runner := mockRunner(&commands, consts.MCPBinaryName)
	installer := mcp.NewInstaller(fetcher, runner)

	ctx := context.Background()
	release := &mcp.ReleaseInfo{
		Tag:       "v0.5.0",
		CommitSHA: sha,
	}

	binaryPath, err := installer.CloneAndBuild(
		ctx, mcp.CloneSSH, release, destDir,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify clone used SSH URL.
	found := false
	for _, cmd := range commands {
		if contains(cmd, consts.GemaraMCPCloneSSH) {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected git clone with SSH URL")
	}

	// Verify checkout used SHA digest.
	foundCheckout := false
	for _, cmd := range commands {
		if contains(cmd, sha) &&
			contains(cmd, "checkout") {
			foundCheckout = true
			break
		}
	}
	if !foundCheckout {
		t.Fatal(
			"expected git checkout with SHA256 digest",
		)
	}

	// Verify binary path exists.
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("binary not found at %s", binaryPath)
	}
}

func TestInstaller_CloneAndBuild_HTTPS(t *testing.T) {
	destDir := t.TempDir()
	var commands []string
	sha := "def456abc123"

	runner := mockRunner(&commands, consts.MCPBinaryName)
	installer := mcp.NewInstaller(
		mockFetcher("v0.5.0", sha), runner,
	)

	ctx := context.Background()
	release := &mcp.ReleaseInfo{
		Tag:       "v0.5.0",
		CommitSHA: sha,
	}

	_, err := installer.CloneAndBuild(
		ctx, mcp.CloneHTTPS, release, destDir,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify clone used HTTPS URL.
	found := false
	for _, cmd := range commands {
		if contains(cmd, consts.GemaraMCPCloneHTTPS) {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected git clone with HTTPS URL")
	}
}

func TestInstaller_InstallPodman(t *testing.T) {
	var commands []string
	runner := func(
		_ context.Context,
		_ string,
		name string,
		args ...string,
	) ([]byte, error) {
		cmd := fmt.Sprintf(
			"%s %v", name, args,
		)
		commands = append(commands, cmd)
		return nil, nil
	}

	installer := mcp.NewInstaller(nil, runner)
	ctx := context.Background()

	if err := installer.InstallPodman(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, cmd := range commands {
		if contains(cmd, "podman") &&
			contains(cmd, "run") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected podman run command")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > 0 && searchString(s, substr))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
