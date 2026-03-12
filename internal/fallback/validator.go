// SPDX-License-Identifier: Apache-2.0

package fallback

import (
	"context"
	"fmt"
	"strings"
)

// CUERunner abstracts cue vet command execution for testing.
type CUERunner func(
	ctx context.Context,
	args ...string,
) ([]byte, error)

// ValidateLocal validates a Gemara artifact against the
// specified schema type using local cue vet. The artifact
// content is passed as a temporary file.
func ValidateLocal(
	ctx context.Context,
	runner CUERunner,
	artifactPath string,
	schemaType string,
	schemaVersion string,
) error {
	definition := schemaType
	if !strings.HasPrefix(definition, "#") {
		definition = "#" + definition
	}

	out, err := runner(
		ctx,
		"vet", "-c",
		"-d", definition,
		artifactPath,
	)
	if err != nil {
		return fmt.Errorf(
			"cue vet failed for %s (schema %s at %s): %s: %w",
			artifactPath,
			definition,
			schemaVersion,
			strings.TrimSpace(string(out)),
			err,
		)
	}
	return nil
}
