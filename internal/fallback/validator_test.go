// SPDX-License-Identifier: Apache-2.0

package fallback_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hbraswelrh/gemara-user-journey/internal/fallback"
)

func TestValidateLocal_Success(t *testing.T) {
	runner := func(
		_ context.Context,
		args ...string,
	) ([]byte, error) {
		// Verify the correct arguments are passed.
		if len(args) < 4 {
			t.Fatalf("expected at least 4 args, got %d", len(args))
		}
		if args[0] != "vet" {
			t.Fatalf("expected 'vet', got %s", args[0])
		}
		if args[1] != "-c" {
			t.Fatalf("expected '-c', got %s", args[1])
		}
		if args[2] != "-d" {
			t.Fatalf("expected '-d', got %s", args[2])
		}
		if args[3] != "#GuidanceCatalog" {
			t.Fatalf(
				"expected '#GuidanceCatalog', got %s",
				args[3],
			)
		}
		return nil, nil
	}

	err := fallback.ValidateLocal(
		context.Background(),
		runner,
		"artifact.yaml",
		"#GuidanceCatalog",
		"v0.20.0",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateLocal_AddsPrefixHash(t *testing.T) {
	var capturedDef string
	runner := func(
		_ context.Context,
		args ...string,
	) ([]byte, error) {
		if len(args) >= 4 {
			capturedDef = args[3]
		}
		return nil, nil
	}

	_ = fallback.ValidateLocal(
		context.Background(),
		runner,
		"artifact.yaml",
		"GuidanceCatalog",
		"v0.20.0",
	)
	if capturedDef != "#GuidanceCatalog" {
		t.Fatalf(
			"expected '#GuidanceCatalog', got %s",
			capturedDef,
		)
	}
}

func TestValidateLocal_CUEVetFails(t *testing.T) {
	runner := func(
		_ context.Context,
		args ...string,
	) ([]byte, error) {
		return []byte(
			"field not allowed: badField",
		), errors.New("exit status 1")
	}

	err := fallback.ValidateLocal(
		context.Background(),
		runner,
		"artifact.yaml",
		"#GuidanceCatalog",
		"v0.20.0",
	)
	if err == nil {
		t.Fatal("expected error from cue vet failure")
	}
}
