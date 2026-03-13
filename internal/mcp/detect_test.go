// SPDX-License-Identifier: Apache-2.0

package mcp_test

import (
	"errors"
	"testing"

	"github.com/hbraswelrh/pacman/internal/mcp"
)

func TestDetect_BinaryFoundInPATH(t *testing.T) {
	lookup := func(name string) (string, error) {
		return "/usr/local/bin/gemara-mcp", nil
	}
	podman := func(container string) (bool, error) {
		return false, nil
	}

	result, err := mcp.Detect(lookup, podman)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Detected {
		t.Fatal("expected Detected to be true")
	}
	if result.Method != mcp.MethodBinary {
		t.Fatalf(
			"expected MethodBinary, got %v",
			result.Method,
		)
	}
	if result.BinaryPath != "/usr/local/bin/gemara-mcp" {
		t.Fatalf(
			"expected binary path, got %q",
			result.BinaryPath,
		)
	}
}

func TestDetect_PodmanContainerRunning(t *testing.T) {
	lookup := func(name string) (string, error) {
		return "", errors.New("not found")
	}
	podman := func(container string) (bool, error) {
		return true, nil
	}

	result, err := mcp.Detect(lookup, podman)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Detected {
		t.Fatal("expected Detected to be true")
	}
	if result.Method != mcp.MethodPodman {
		t.Fatalf(
			"expected MethodPodman, got %v",
			result.Method,
		)
	}
}

func TestDetect_NeitherBinaryNorPodman(t *testing.T) {
	lookup := func(name string) (string, error) {
		return "", errors.New("not found")
	}
	podman := func(container string) (bool, error) {
		return false, nil
	}

	result, err := mcp.Detect(lookup, podman)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Detected {
		t.Fatal("expected Detected to be false")
	}
	if result.Method != mcp.MethodNone {
		t.Fatalf(
			"expected MethodNone, got %v",
			result.Method,
		)
	}
}

func TestDetect_PodmanCheckError(t *testing.T) {
	lookup := func(name string) (string, error) {
		return "", errors.New("not found")
	}
	podman := func(container string) (bool, error) {
		return false, errors.New("podman not running")
	}

	_, err := mcp.Detect(lookup, podman)
	if err == nil {
		t.Fatal("expected error from Podman check")
	}
}
