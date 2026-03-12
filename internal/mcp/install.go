// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// CloneMethod specifies how to clone the gemara-mcp repository.
type CloneMethod int

const (
	// CloneSSH clones via SSH.
	CloneSSH CloneMethod = iota
	// CloneHTTPS clones via HTTPS.
	CloneHTTPS
)

// ReleaseInfo holds information about a gemara-mcp release.
type ReleaseInfo struct {
	// Tag is the release tag (e.g., "v0.5.0").
	Tag string
	// CommitSHA is the SHA256 commit digest for the release.
	CommitSHA string
}

// GitHubRelease represents a subset of the GitHub API release
// response.
type GitHubRelease struct {
	TagName   string `json:"tag_name"`
	TargetSHA string `json:"target_commitish"`
}

// ReleaseFetcher abstracts GitHub release API calls for
// testing.
type ReleaseFetcher func(
	ctx context.Context,
	repoURL string,
) (*ReleaseInfo, error)

// CommandRunner abstracts shell command execution for testing.
type CommandRunner func(
	ctx context.Context,
	dir string,
	name string,
	args ...string,
) ([]byte, error)

// DefaultReleaseFetcher queries the GitHub API for the latest
// release of the given repository and returns the tag and
// commit SHA.
func DefaultReleaseFetcher(
	ctx context.Context,
	repoURL string,
) (*ReleaseInfo, error) {
	apiURL := strings.Replace(
		repoURL,
		"https://github.com/",
		"https://api.github.com/repos/",
		1,
	) + "/releases/latest"

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, apiURL, nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create request: %w", err,
		)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"GitHub API returned %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(
		&release,
	); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &ReleaseInfo{
		Tag:       release.TagName,
		CommitSHA: release.TargetSHA,
	}, nil
}

// DefaultCommandRunner executes a command in the given
// directory and returns its combined output.
func DefaultCommandRunner(
	ctx context.Context,
	dir string,
	name string,
	args ...string,
) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// Installer handles the automated gemara-mcp installation
// pipeline.
type Installer struct {
	fetchRelease ReleaseFetcher
	runCommand   CommandRunner
}

// NewInstaller creates an Installer with the given
// dependencies.
func NewInstaller(
	fetcher ReleaseFetcher,
	runner CommandRunner,
) *Installer {
	return &Installer{
		fetchRelease: fetcher,
		runCommand:   runner,
	}
}

// ResolveLatestRelease fetches the latest gemara-mcp release
// and returns its tag and SHA256 commit digest.
func (i *Installer) ResolveLatestRelease(
	ctx context.Context,
) (*ReleaseInfo, error) {
	return i.fetchRelease(ctx, consts.GemaraMCPRepoURL)
}

// CloneAndBuild clones the gemara-mcp repository at the
// specified commit digest and runs make build.
func (i *Installer) CloneAndBuild(
	ctx context.Context,
	method CloneMethod,
	release *ReleaseInfo,
	destDir string,
) (string, error) {
	cloneURL := consts.GemaraMCPCloneHTTPS
	if method == CloneSSH {
		cloneURL = consts.GemaraMCPCloneSSH
	}

	// Clone the repository.
	repoDir := filepath.Join(destDir, consts.MCPBinaryName)
	_, err := i.runCommand(
		ctx, "",
		"git", "clone", cloneURL, repoDir,
	)
	if err != nil {
		return "", fmt.Errorf("git clone: %w", err)
	}

	// Checkout the pinned commit by SHA256 digest.
	_, err = i.runCommand(
		ctx, repoDir,
		"git", "checkout", release.CommitSHA,
	)
	if err != nil {
		return "", fmt.Errorf("git checkout: %w", err)
	}

	// Build the binary.
	_, err = i.runCommand(ctx, repoDir, "make", "build")
	if err != nil {
		return "", fmt.Errorf("make build: %w", err)
	}

	binaryPath := filepath.Join(repoDir, consts.MCPBinaryName)
	if _, err := os.Stat(binaryPath); err != nil {
		return "", fmt.Errorf(
			"binary not found at %s: %w",
			binaryPath,
			err,
		)
	}

	return binaryPath, nil
}

// InstallPodman provides the Podman run configuration for the
// gemara-mcp server and returns the container name.
func (i *Installer) InstallPodman(
	ctx context.Context,
) error {
	_, err := i.runCommand(
		ctx, "",
		"podman", "run", "-d",
		"--name", consts.MCPPodmanContainer,
		consts.MCPPodmanImage,
	)
	if err != nil {
		return fmt.Errorf("podman run: %w", err)
	}
	return nil
}
