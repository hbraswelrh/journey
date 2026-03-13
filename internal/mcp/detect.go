// SPDX-License-Identifier: Apache-2.0

// Package mcp provides detection, installation, client
// connectivity, and version compatibility for the Gemara MCP
// server (gemara-mcp).
package mcp

import (
	"os/exec"
	"strings"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// InstallMethod indicates how the MCP server was installed.
type InstallMethod int

const (
	// MethodNone means the MCP server was not detected.
	MethodNone InstallMethod = iota
	// MethodBinary means the MCP server binary is in PATH.
	MethodBinary
	// MethodPodman means the MCP server is running in Podman.
	MethodPodman
)

// DetectionResult holds the outcome of MCP server detection.
type DetectionResult struct {
	// Detected is true if the MCP server was found.
	Detected bool
	// Method indicates how the server was found.
	Method InstallMethod
	// BinaryPath is the resolved path to the binary (if
	// Method == MethodBinary).
	BinaryPath string
}

// BinaryLookup abstracts PATH-based binary lookup for testing.
type BinaryLookup func(name string) (string, error)

// PodmanChecker abstracts Podman container status checks for
// testing.
type PodmanChecker func(container string) (bool, error)

// DefaultBinaryLookup uses exec.LookPath to find a binary.
func DefaultBinaryLookup(name string) (string, error) {
	return exec.LookPath(name)
}

// DefaultPodmanChecker uses podman inspect to check if a
// container is running.
func DefaultPodmanChecker(
	container string,
) (bool, error) {
	out, err := exec.Command(
		"podman", "inspect",
		"--format", "{{.State.Running}}",
		container,
	).Output()
	if err != nil {
		return false, nil //nolint:nilerr // not found is ok
	}
	return strings.TrimSpace(string(out)) == "true", nil
}

// Detect checks whether the Gemara MCP server is installed and
// accessible. It first checks for the binary in PATH, then
// checks for a running Podman container.
func Detect(
	lookupBinary BinaryLookup,
	checkPodman PodmanChecker,
) (DetectionResult, error) {
	// Check for binary in PATH.
	path, err := lookupBinary(consts.MCPBinaryName)
	if err == nil && path != "" {
		return DetectionResult{
			Detected:   true,
			Method:     MethodBinary,
			BinaryPath: path,
		}, nil
	}

	// Check for Podman container.
	running, err := checkPodman(consts.MCPPodmanContainer)
	if err != nil {
		return DetectionResult{}, err
	}
	if running {
		return DetectionResult{
			Detected: true,
			Method:   MethodPodman,
		}, nil
	}

	return DetectionResult{
		Detected: false,
		Method:   MethodNone,
	}, nil
}
