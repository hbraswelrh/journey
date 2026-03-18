// SPDX-License-Identifier: Apache-2.0

package tutorials

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// FetchConfig holds configuration for fetching tutorials
// from the upstream Gemara repository.
type FetchConfig struct {
	// HomeDir is the user's home directory.
	HomeDir string
	// GemaraDir overrides the default clone location.
	// If empty, uses HomeDir/DefaultGemaraDir.
	GemaraDir string
	// UseSSH uses SSH clone URL instead of HTTPS.
	UseSSH bool
}

// FetchResult holds the outcome of a tutorial fetch.
type FetchResult struct {
	// TutorialsDir is the resolved path to the tutorials.
	TutorialsDir string
	// Cloned is true if a fresh clone was performed.
	Cloned bool
	// Updated is true if an existing clone was pulled.
	Updated bool
	// Branch is the checked-out branch.
	Branch string
}

// ResolveTutorialsDir returns the path to the tutorials
// directory, cloning or updating the upstream Gemara
// repository as needed.
func ResolveTutorialsDir(
	cfg *FetchConfig,
) (*FetchResult, error) {
	gemaraDir := cfg.GemaraDir
	if gemaraDir == "" {
		gemaraDir = filepath.Join(
			cfg.HomeDir, consts.DefaultGemaraDir,
		)
	}

	tutorialsDir := filepath.Join(
		gemaraDir, consts.GemaraTutorialsSubdir,
	)

	// Check if the repo is already cloned.
	if isGitRepo(gemaraDir) {
		// Pull latest from main.
		if err := gitPull(gemaraDir); err != nil {
			// Pull failed — use existing content.
			return &FetchResult{
				TutorialsDir: tutorialsDir,
				Updated:      false,
				Branch:       "main",
			}, nil
		}
		return &FetchResult{
			TutorialsDir: tutorialsDir,
			Updated:      true,
			Branch:       "main",
		}, nil
	}

	// Check if the tutorials dir exists (user may have
	// cloned the repo manually elsewhere).
	if dirExists(tutorialsDir) {
		return &FetchResult{
			TutorialsDir: tutorialsDir,
		}, nil
	}

	// Clone the repository.
	cloneURL := consts.GemaraCloneHTTPS
	if cfg.UseSSH {
		cloneURL = consts.GemaraCloneSSH
	}

	if err := gitClone(
		cloneURL, gemaraDir,
	); err != nil {
		return nil, fmt.Errorf(
			"clone gemara repository: %w", err,
		)
	}

	return &FetchResult{
		TutorialsDir: tutorialsDir,
		Cloned:       true,
		Branch:       "main",
	}, nil
}

// isGitRepo checks if the directory is a git repository.
func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// dirExists checks if a directory exists.
func dirExists(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// gitClone clones a repository to the given directory,
// checking out the main branch.
func gitClone(url, dest string) error {
	// Ensure parent directory exists.
	parent := filepath.Dir(dest)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	cmd := exec.Command(
		"git", "clone",
		"--branch", "main",
		"--single-branch",
		"--depth", "1",
		url, dest,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// gitPull pulls the latest changes on the main branch.
func gitPull(dir string) error {
	// Ensure we're on main.
	checkout := exec.Command(
		"git", "checkout", "main",
	)
	checkout.Dir = dir
	checkout.Stdout = os.Stdout
	checkout.Stderr = os.Stderr
	if err := checkout.Run(); err != nil {
		return err
	}

	pull := exec.Command(
		"git", "pull", "origin", "main",
	)
	pull.Dir = dir
	pull.Stdout = os.Stdout
	pull.Stderr = os.Stderr
	return pull.Run()
}

// ExpandTutorialsDir resolves ~ in the tutorials path
// and checks common locations for the Gemara tutorials.
// Returns the first valid tutorials directory found, or
// attempts to fetch from upstream if none exists.
func ExpandTutorialsDir(
	dir string,
	homeDir string,
) string {
	// Expand ~ prefix.
	if strings.HasPrefix(dir, "~/") {
		dir = filepath.Join(
			homeDir, dir[2:],
		)
	}

	// Check if the expanded path exists.
	if dirExists(dir) {
		return dir
	}

	// Check the managed clone location.
	managed := filepath.Join(
		homeDir,
		consts.DefaultGemaraDir,
		consts.GemaraTutorialsSubdir,
	)
	if dirExists(managed) {
		return managed
	}

	// Return the original (will trigger a fetch or
	// error in the caller).
	return dir
}
