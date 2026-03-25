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

	"github.com/hbraswelrh/journey/internal/consts"
)

// CloneMethod specifies how to clone the gemara-mcp repository.
type CloneMethod int

const (
	// CloneSSH clones via SSH.
	CloneSSH CloneMethod = iota
	// CloneHTTPS clones via HTTPS.
	CloneHTTPS
)

// SSHChecker abstracts SSH key detection for testing.
type SSHChecker func(ctx context.Context) bool

// DefaultSSHChecker probes GitHub SSH access by running
// `ssh -T git@github.com`. GitHub returns exit code 1 with
// "successfully authenticated" when SSH keys are configured,
// or a connection/auth error when they are not.
func DefaultSSHChecker(ctx context.Context) bool {
	out, _ := exec.CommandContext(
		ctx, "ssh",
		"-T",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=5",
		"git@github.com",
	).CombinedOutput()
	return strings.Contains(
		string(out), "successfully authenticated",
	)
}

// DetectCloneMethod returns CloneSSH if the user has SSH
// keys configured for GitHub, or CloneHTTPS otherwise.
// HTTPS is always the safe default.
func DetectCloneMethod(
	ctx context.Context,
	checker SSHChecker,
) CloneMethod {
	if checker(ctx) {
		return CloneSSH
	}
	return CloneHTTPS
}

// CloneMethodLabel returns a human-readable label for the
// clone method.
func CloneMethodLabel(m CloneMethod) string {
	if m == CloneSSH {
		return "SSH"
	}
	return "HTTPS"
}

// ReleaseInfo holds information about a gemara-mcp release.
type ReleaseInfo struct {
	// Tag is the release tag (e.g., "v0.5.0").
	Tag string
	// CommitSHA is the SHA256 commit digest for the release.
	CommitSHA string
	// Prerelease indicates this is not a stable release.
	Prerelease bool
}

// GitHubRelease represents a subset of the GitHub API release
// response.
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	TargetSHA  string `json:"target_commitish"`
	Prerelease bool   `json:"prerelease"`
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
// commit SHA. If no stable release exists, it falls back to
// the most recent prerelease so users can test with
// development builds until official releases are published.
func DefaultReleaseFetcher(
	ctx context.Context,
	repoURL string,
) (*ReleaseInfo, error) {
	apiBase := strings.Replace(
		repoURL,
		"https://github.com/",
		"https://api.github.com/repos/",
		1,
	)

	// Try the stable /releases/latest endpoint first.
	info, err := fetchRelease(
		ctx, apiBase+"/releases/latest",
	)
	if err == nil {
		return info, nil
	}

	// If no stable release, fall back to the releases
	// list (includes prereleases).
	return fetchLatestFromList(
		ctx, apiBase+"/releases?per_page=1",
	)
}

// fetchRelease fetches a single release from the GitHub API.
func fetchRelease(
	ctx context.Context,
	url string,
) (*ReleaseInfo, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, url, nil,
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

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no release found")
	}

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
		Tag:        release.TagName,
		CommitSHA:  release.TargetSHA,
		Prerelease: release.Prerelease,
	}, nil
}

// fetchLatestFromList fetches the most recent release
// (including prereleases) from the GitHub releases list.
func fetchLatestFromList(
	ctx context.Context,
	url string,
) (*ReleaseInfo, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, url, nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create request: %w", err,
		)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
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

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(
		&releases,
	); err != nil {
		return nil, fmt.Errorf(
			"decode releases: %w", err,
		)
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf(
			"the gemara-mcp repository has no " +
				"published releases yet — skip " +
				"installation and try again later",
		)
	}

	r := releases[0]
	return &ReleaseInfo{
		Tag:        r.TagName,
		CommitSHA:  r.TargetSHA,
		Prerelease: r.Prerelease,
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
// specified commit digest and runs make build. If the
// repository was previously cloned, it fetches updates and
// checks out the requested SHA instead of cloning again.
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

	repoDir := filepath.Join(destDir, consts.MCPBinaryName)

	if isGitRepo(repoDir) {
		// Repository already exists — fetch and checkout
		// the pinned SHA.
		out, err := i.runCommand(
			ctx, repoDir,
			"git", "fetch", "--all",
		)
		if err != nil {
			return "", fmt.Errorf(
				"git fetch: %s: %w",
				string(out), err,
			)
		}
	} else {
		// Ensure parent directory exists.
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return "", fmt.Errorf(
				"create install dir: %w", err,
			)
		}

		// Remove any leftover non-git directory (e.g.,
		// from a failed prior install).
		if _, err := os.Stat(repoDir); err == nil {
			if err := os.RemoveAll(repoDir); err != nil {
				return "", fmt.Errorf(
					"clean prior install dir: %w", err,
				)
			}
		}

		// Fresh clone.
		out, err := i.runCommand(
			ctx, "",
			"git", "clone", cloneURL, repoDir,
		)
		if err != nil {
			return "", fmt.Errorf(
				"git clone: %s: %w",
				string(out), err,
			)
		}
	}

	// Checkout the pinned commit by SHA digest.
	out, err := i.runCommand(
		ctx, repoDir,
		"git", "checkout", release.CommitSHA,
	)
	if err != nil {
		return "", fmt.Errorf(
			"git checkout %s: %s: %w",
			truncateForLog(release.CommitSHA),
			string(out), err,
		)
	}

	// Build the binary.
	out, err = i.runCommand(
		ctx, repoDir, "make", "build",
	)
	if err != nil {
		return "", fmt.Errorf(
			"make build: %s: %w",
			string(out), err,
		)
	}

	binaryPath := filepath.Join(
		repoDir, "bin", consts.MCPBinaryName,
	)
	if _, err := os.Stat(binaryPath); err != nil {
		// Fall back to root-level binary for repos that
		// build to the project root.
		binaryPath = filepath.Join(
			repoDir, consts.MCPBinaryName,
		)
		if _, err := os.Stat(binaryPath); err != nil {
			return "", fmt.Errorf(
				"binary not found after build: %w",
				err,
			)
		}
	}

	return binaryPath, nil
}

// isGitRepo returns true if the directory exists and
// contains a .git subdirectory.
func isGitRepo(dir string) bool {
	info, err := os.Stat(
		filepath.Join(dir, ".git"),
	)
	return err == nil && info.IsDir()
}

// truncateForLog truncates a string to 12 characters for
// log output.
func truncateForLog(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
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

// InstalledRelease records the release metadata for a
// source-built installation. This file is stored alongside
// the built binary so update checks can compare against
// the latest upstream release.
type InstalledRelease struct {
	// Tag is the release tag (e.g., "v0.0.0").
	Tag string `json:"tag"`
	// CommitSHA is the pinned SHA digest of the installed
	// commit.
	CommitSHA string `json:"commit_sha"`
	// Prerelease indicates whether this was a prerelease.
	Prerelease bool `json:"prerelease"`
	// InstalledAt is the ISO 8601 timestamp of
	// installation.
	InstalledAt string `json:"installed_at"`
	// BinaryPath is the path to the built binary.
	BinaryPath string `json:"binary_path"`
}

// SaveInstalledRelease writes the installed release metadata
// to the install directory.
func SaveInstalledRelease(
	installDir string,
	release *InstalledRelease,
) error {
	data, err := json.MarshalIndent(release, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal release: %w", err)
	}
	path := filepath.Join(
		installDir, consts.InstalledReleaseFile,
	)
	if err := os.WriteFile(
		path, data, 0o644,
	); err != nil {
		return fmt.Errorf(
			"write installed release: %w", err,
		)
	}
	return nil
}

// LoadInstalledRelease reads the installed release metadata
// from the install directory. Returns nil if no metadata file
// exists (first install or non-source installation).
func LoadInstalledRelease(
	installDir string,
) (*InstalledRelease, error) {
	path := filepath.Join(
		installDir, consts.InstalledReleaseFile,
	)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf(
			"read installed release: %w", err,
		)
	}
	var release InstalledRelease
	if err := json.Unmarshal(data, &release); err != nil {
		return nil, fmt.Errorf(
			"parse installed release: %w", err,
		)
	}
	return &release, nil
}

// UpdateCheck holds the result of comparing the installed
// release against the latest upstream release.
type UpdateCheck struct {
	// UpdateAvailable is true when the upstream SHA differs
	// from the installed SHA.
	UpdateAvailable bool
	// Installed is the currently installed release metadata.
	Installed *InstalledRelease
	// Latest is the latest upstream release.
	Latest *ReleaseInfo
}

// CheckForUpdate compares the locally installed release
// against the latest upstream release. Returns an UpdateCheck
// indicating whether an update is available.
func (i *Installer) CheckForUpdate(
	ctx context.Context,
	installDir string,
) (*UpdateCheck, error) {
	installed, err := LoadInstalledRelease(installDir)
	if err != nil {
		return nil, fmt.Errorf(
			"load installed release: %w", err,
		)
	}
	if installed == nil {
		// No metadata file — cannot determine installed
		// version (e.g., first install or Podman-based).
		return &UpdateCheck{
			UpdateAvailable: false,
		}, nil
	}

	latest, err := i.ResolveLatestRelease(ctx)
	if err != nil {
		// Network error — skip update check silently.
		return &UpdateCheck{
			UpdateAvailable: false,
			Installed:       installed,
		}, nil
	}

	return &UpdateCheck{
		UpdateAvailable: installed.CommitSHA !=
			latest.CommitSHA,
		Installed: installed,
		Latest:    latest,
	}, nil
}
