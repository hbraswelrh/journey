// SPDX-License-Identifier: Apache-2.0

package mcp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hbraswelrh/journey/internal/mcp"
)

func versionFetcher(
	serverVer, schemaVer string,
) mcp.VersionFetcher {
	return func(
		_ context.Context,
		_ *mcp.Client,
	) (*mcp.VersionInfo, error) {
		return &mcp.VersionInfo{
			ServerVersion: serverVer,
			SchemaVersion: schemaVer,
		}, nil
	}
}

func failingVersionFetcher() mcp.VersionFetcher {
	return func(
		_ context.Context,
		_ *mcp.Client,
	) (*mcp.VersionInfo, error) {
		return nil, errors.New("not supported")
	}
}

func TestCheckCompatibility_VersionsMatch(t *testing.T) {
	fetcher := versionFetcher("v0.5.0", "v0.20.0")

	result, err := mcp.CheckCompatibility(
		context.Background(),
		fetcher,
		nil,
		"v0.20.0",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != mcp.CompatOK {
		t.Fatalf(
			"expected CompatOK, got %v",
			result.Status,
		)
	}
	if result.Recommendation != "" {
		t.Fatalf(
			"expected no recommendation, got %q",
			result.Recommendation,
		)
	}
}

func TestCheckCompatibility_VersionsMismatch(t *testing.T) {
	fetcher := versionFetcher("v0.4.0", "v0.18.0")

	result, err := mcp.CheckCompatibility(
		context.Background(),
		fetcher,
		nil,
		"v0.20.0",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != mcp.CompatMismatch {
		t.Fatalf(
			"expected CompatMismatch, got %v",
			result.Status,
		)
	}
	if result.Recommendation == "" {
		t.Fatal("expected non-empty recommendation")
	}
	if result.MCPSchemaVersion != "v0.18.0" {
		t.Fatalf(
			"expected MCPSchemaVersion v0.18.0, got %s",
			result.MCPSchemaVersion,
		)
	}
}

func TestCheckCompatibility_NoVersionMetadata(t *testing.T) {
	fetcher := failingVersionFetcher()

	result, err := mcp.CheckCompatibility(
		context.Background(),
		fetcher,
		nil,
		"v0.20.0",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != mcp.CompatUnknown {
		t.Fatalf(
			"expected CompatUnknown, got %v",
			result.Status,
		)
	}
	if result.Recommendation == "" {
		t.Fatal("expected non-empty recommendation")
	}
}
