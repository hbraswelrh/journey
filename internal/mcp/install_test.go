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

func TestDefaultReleaseFetcher_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Call against a repo that truly does not exist.
	_, err := mcp.DefaultReleaseFetcher(
		ctx,
		"https://github.com/gemaraproj/"+
			"definitely-nonexistent-repo-xyz",
	)
	if err == nil {
		t.Fatal("expected error for nonexistent repo")
	}
	// Should get the user-friendly message, not raw JSON.
	errMsg := err.Error()
	if !searchString(errMsg, "no published releases") &&
		!searchString(errMsg, "no release found") &&
		!searchString(errMsg, "404") {
		t.Errorf(
			"expected error for missing repo, "+
				"got: %s",
			errMsg,
		)
	}
}

// TestDefaultReleaseFetcher_PrereleasesFallback verifies
// that when no stable release exists but prereleases do,
// the fetcher returns the prerelease.
func TestDefaultReleaseFetcher_Prerelease(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// The real gemara-mcp repo has v0.0.0 prerelease
	// but no stable release — this should succeed via
	// the fallback path.
	info, err := mcp.DefaultReleaseFetcher(
		ctx,
		"https://github.com/gemaraproj/gemara-mcp",
	)
	if err != nil {
		t.Fatalf("expected prerelease fallback, got: %v", err)
	}
	if info.Tag == "" {
		t.Error("expected non-empty tag")
	}
	if info.CommitSHA == "" {
		t.Error("expected non-empty CommitSHA")
	}
	if !info.Prerelease {
		t.Log(
			"release is not marked as prerelease " +
				"— a stable release may now exist",
		)
	}
}

// TestSaveAndLoadInstalledRelease verifies round-trip of
// installed release metadata.
func TestSaveAndLoadInstalledRelease(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	release := &mcp.InstalledRelease{
		Tag:         "v0.0.0",
		CommitSHA:   "015fcbc483ff18e48fc1063cb8f9e35298a6830c",
		Prerelease:  true,
		InstalledAt: "2026-03-13T12:00:00Z",
		BinaryPath:  "/usr/local/bin/gemara-mcp",
	}

	err := mcp.SaveInstalledRelease(dir, release)
	if err != nil {
		t.Fatalf("SaveInstalledRelease: %v", err)
	}

	loaded, err := mcp.LoadInstalledRelease(dir)
	if err != nil {
		t.Fatalf("LoadInstalledRelease: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil loaded release")
	}
	if loaded.Tag != release.Tag {
		t.Errorf(
			"Tag = %q, want %q",
			loaded.Tag, release.Tag,
		)
	}
	if loaded.CommitSHA != release.CommitSHA {
		t.Errorf(
			"CommitSHA = %q, want %q",
			loaded.CommitSHA, release.CommitSHA,
		)
	}
	if !loaded.Prerelease {
		t.Error("expected Prerelease = true")
	}
}

// TestLoadInstalledRelease_NoFile returns nil when no
// metadata file exists.
func TestLoadInstalledRelease_NoFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	loaded, err := mcp.LoadInstalledRelease(dir)
	if err != nil {
		t.Fatalf("LoadInstalledRelease: %v", err)
	}
	if loaded != nil {
		t.Error("expected nil when no file exists")
	}
}

// TestCheckForUpdate_NoUpdate returns false when installed
// SHA matches latest.
func TestCheckForUpdate_NoUpdate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sha := "abc123def456abc123def456"

	// Save installed release.
	release := &mcp.InstalledRelease{
		Tag:         "v0.5.0",
		CommitSHA:   sha,
		InstalledAt: "2026-03-13T12:00:00Z",
		BinaryPath:  "/usr/local/bin/gemara-mcp",
	}
	_ = mcp.SaveInstalledRelease(dir, release)

	// Fetcher returns same SHA.
	fetcher := func(
		_ context.Context,
		_ string,
	) (*mcp.ReleaseInfo, error) {
		return &mcp.ReleaseInfo{
			Tag:       "v0.5.0",
			CommitSHA: sha,
		}, nil
	}

	installer := mcp.NewInstaller(fetcher, nil)
	update, err := installer.CheckForUpdate(
		context.Background(), dir,
	)
	if err != nil {
		t.Fatalf("CheckForUpdate: %v", err)
	}
	if update.UpdateAvailable {
		t.Error("expected no update available")
	}
}

// TestCheckForUpdate_UpdateAvailable returns true when SHAs
// differ.
func TestCheckForUpdate_UpdateAvailable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Save installed release with old SHA.
	release := &mcp.InstalledRelease{
		Tag:         "v0.5.0",
		CommitSHA:   "oldsha123",
		InstalledAt: "2026-03-13T12:00:00Z",
		BinaryPath:  "/usr/local/bin/gemara-mcp",
	}
	_ = mcp.SaveInstalledRelease(dir, release)

	// Fetcher returns new SHA.
	fetcher := func(
		_ context.Context,
		_ string,
	) (*mcp.ReleaseInfo, error) {
		return &mcp.ReleaseInfo{
			Tag:       "v0.6.0",
			CommitSHA: "newsha456",
		}, nil
	}

	installer := mcp.NewInstaller(fetcher, nil)
	update, err := installer.CheckForUpdate(
		context.Background(), dir,
	)
	if err != nil {
		t.Fatalf("CheckForUpdate: %v", err)
	}
	if !update.UpdateAvailable {
		t.Error("expected update available")
	}
	if update.Installed.Tag != "v0.5.0" {
		t.Errorf(
			"Installed.Tag = %q, want %q",
			update.Installed.Tag, "v0.5.0",
		)
	}
	if update.Latest.Tag != "v0.6.0" {
		t.Errorf(
			"Latest.Tag = %q, want %q",
			update.Latest.Tag, "v0.6.0",
		)
	}
}

// TestCheckForUpdate_NoMetadata returns no update when
// metadata file does not exist.
func TestCheckForUpdate_NoMetadata(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	fetcher := func(
		_ context.Context,
		_ string,
	) (*mcp.ReleaseInfo, error) {
		return &mcp.ReleaseInfo{
			Tag:       "v0.6.0",
			CommitSHA: "newsha456",
		}, nil
	}

	installer := mcp.NewInstaller(fetcher, nil)
	update, err := installer.CheckForUpdate(
		context.Background(), dir,
	)
	if err != nil {
		t.Fatalf("CheckForUpdate: %v", err)
	}
	if update.UpdateAvailable {
		t.Error(
			"expected no update when no metadata " +
				"exists",
		)
	}
}

// TestDetectCloneMethod_SSHAvailable returns CloneSSH.
func TestDetectCloneMethod_SSHAvailable(t *testing.T) {
	t.Parallel()
	checker := func(_ context.Context) bool {
		return true
	}
	method := mcp.DetectCloneMethod(
		context.Background(), checker,
	)
	if method != mcp.CloneSSH {
		t.Errorf("expected CloneSSH, got %d", method)
	}
}

// TestDetectCloneMethod_SSHUnavailable returns CloneHTTPS.
func TestDetectCloneMethod_SSHUnavailable(t *testing.T) {
	t.Parallel()
	checker := func(_ context.Context) bool {
		return false
	}
	method := mcp.DetectCloneMethod(
		context.Background(), checker,
	)
	if method != mcp.CloneHTTPS {
		t.Errorf("expected CloneHTTPS, got %d", method)
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
