// SPDX-License-Identifier: Apache-2.0

package authoring

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// errCUEFailed is a sentinel error indicating cue vet
// returned a non-zero exit code.
var errCUEFailed = errors.New("cue vet validation failed")

// MCPClient is the subset of the MCP client interface used
// by the validator.
type MCPClient interface {
	ValidateArtifact(
		ctx context.Context,
		artifact string,
		schemaType string,
	) ([]byte, error)
}

// CUERunner abstracts cue vet command execution.
type CUERunner func(
	ctx context.Context,
	args ...string,
) ([]byte, error)

// Validator validates authored artifacts against the Gemara
// CUE schema.
type Validator interface {
	// ValidatePartial validates the artifact up to the
	// given step index.
	ValidatePartial(
		ctx context.Context,
		artifact *AuthoredArtifact,
		stepIdx int,
	) ([]ValidationError, error)
	// ValidateFull validates the complete artifact.
	ValidateFull(
		ctx context.Context,
		artifact *AuthoredArtifact,
	) ([]ValidationError, error)
}

// MCPValidator validates artifacts using the Gemara MCP
// server's validate_gemara_artifact tool.
type MCPValidator struct {
	client MCPClient
}

// NewMCPValidator creates a validator that uses the MCP
// server.
func NewMCPValidator(client MCPClient) *MCPValidator {
	return &MCPValidator{client: client}
}

// mcpResponse represents the validation response from the
// MCP server.
type mcpResponse struct {
	Valid  bool       `json:"valid"`
	Errors []mcpError `json:"errors,omitempty"`
}

// mcpError represents a single validation error from the
// MCP server.
type mcpError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// ValidatePartial validates the artifact up to the given
// step using the MCP server.
func (v *MCPValidator) ValidatePartial(
	ctx context.Context,
	artifact *AuthoredArtifact,
	_ int,
) ([]ValidationError, error) {
	return v.validate(ctx, artifact)
}

// ValidateFull validates the complete artifact using the
// MCP server.
func (v *MCPValidator) ValidateFull(
	ctx context.Context,
	artifact *AuthoredArtifact,
) ([]ValidationError, error) {
	return v.validate(ctx, artifact)
}

func (v *MCPValidator) validate(
	ctx context.Context,
	artifact *AuthoredArtifact,
) ([]ValidationError, error) {
	content, err := serializeArtifact(artifact)
	if err != nil {
		return nil, fmt.Errorf(
			"serialize artifact: %w", err,
		)
	}

	resp, err := v.client.ValidateArtifact(
		ctx, string(content), artifact.SchemaDef,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"MCP validate: %w", err,
		)
	}

	return parseMCPResponse(resp)
}

// parseMCPResponse parses the MCP server's validation
// response into ValidationError entries.
func parseMCPResponse(
	data []byte,
) ([]ValidationError, error) {
	var resp mcpResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf(
			"parse MCP response: %w", err,
		)
	}
	if resp.Valid {
		return nil, nil
	}
	var errs []ValidationError
	for _, e := range resp.Errors {
		errs = append(errs, ValidationError{
			FieldPath: e.Path,
			Message:   e.Message,
			FixSuggestion: fmt.Sprintf(
				"Check the value at %q and ensure "+
					"it conforms to the schema",
				e.Path,
			),
		})
	}
	return errs, nil
}

// LocalValidator validates artifacts using local cue vet.
type LocalValidator struct {
	runner CUERunner
}

// NewLocalValidator creates a validator that uses local CUE
// tooling.
func NewLocalValidator(runner CUERunner) *LocalValidator {
	return &LocalValidator{runner: runner}
}

// ValidatePartial validates the artifact up to the given
// step using local cue vet.
func (v *LocalValidator) ValidatePartial(
	ctx context.Context,
	artifact *AuthoredArtifact,
	_ int,
) ([]ValidationError, error) {
	return v.validate(ctx, artifact)
}

// ValidateFull validates the complete artifact using local
// cue vet.
func (v *LocalValidator) ValidateFull(
	ctx context.Context,
	artifact *AuthoredArtifact,
) ([]ValidationError, error) {
	return v.validate(ctx, artifact)
}

func (v *LocalValidator) validate(
	ctx context.Context,
	artifact *AuthoredArtifact,
) ([]ValidationError, error) {
	content, err := serializeArtifact(artifact)
	if err != nil {
		return nil, fmt.Errorf(
			"serialize artifact: %w", err,
		)
	}

	// Write to a temp file for cue vet.
	tmpFile, err := os.CreateTemp("", "journey-*.yaml")
	if err != nil {
		return nil, fmt.Errorf(
			"create temp file: %w", err,
		)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf(
			"write temp file: %w", err,
		)
	}
	tmpFile.Close()

	definition := artifact.SchemaDef
	if !strings.HasPrefix(definition, "#") {
		definition = "#" + definition
	}

	out, err := v.runner(
		ctx,
		"vet", "-c",
		"-d", definition,
		tmpFile.Name(),
	)
	if err != nil {
		// Parse cue vet errors.
		return parseValidationOutput(
			string(out),
		), nil
	}
	return nil, nil
}

// parseValidationOutput parses cue vet error output into
// structured ValidationError entries.
func parseValidationOutput(
	output string,
) []ValidationError {
	if strings.TrimSpace(output) == "" {
		return []ValidationError{
			{
				FieldPath: "artifact",
				Message:   "validation failed",
				FixSuggestion: "Review all required " +
					"fields and ensure they conform " +
					"to the schema",
			},
		}
	}

	var errs []ValidationError
	lines := strings.Split(
		strings.TrimSpace(output), "\n",
	)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ve := ValidationError{
			Message: line,
			FixSuggestion: "Review and correct " +
				"the value to conform to the " +
				"Gemara schema",
		}
		// Try to extract field path from cue output.
		// Format: "path.to.field: error message"
		if idx := strings.Index(line, ":"); idx > 0 {
			ve.FieldPath = strings.TrimSpace(
				line[:idx],
			)
			ve.Message = strings.TrimSpace(
				line[idx+1:],
			)
			ve.FixSuggestion = fmt.Sprintf(
				"Check the value at %q and ensure "+
					"it conforms to the Gemara schema",
				ve.FieldPath,
			)
		}
		errs = append(errs, ve)
	}
	return errs
}

// NewValidator returns the appropriate validator based on
// MCP availability.
func NewValidator(
	mcpAvailable bool,
	mcpClient MCPClient,
	runner CUERunner,
) Validator {
	if mcpAvailable && mcpClient != nil {
		return NewMCPValidator(mcpClient)
	}
	return NewLocalValidator(runner)
}

// serializeArtifact converts an authored artifact into YAML
// for validation.
func serializeArtifact(
	artifact *AuthoredArtifact,
) ([]byte, error) {
	// Build a map structure from sections.
	data := make(map[string]interface{})
	for _, section := range artifact.Sections {
		sectionData := make(map[string]interface{})
		for k, v := range section.Fields {
			sectionData[k] = v
		}
		data[section.Name] = sectionData
	}
	return yaml.Marshal(data)
}
