// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// ErrNotConnected is returned when an operation is attempted
// on a client that is not connected.
var ErrNotConnected = errors.New("mcp client: not connected")

// ErrHealthCheckFailed is returned when the MCP server does
// not respond to a health check within the timeout.
var ErrHealthCheckFailed = errors.New(
	"mcp client: health check failed",
)

// ConnectionStatus represents the MCP client connection state.
type ConnectionStatus int

const (
	// StatusDisconnected means the client is not connected.
	StatusDisconnected ConnectionStatus = iota
	// StatusConnected means the client has an active
	// connection.
	StatusConnected
)

// Transport abstracts the MCP protocol communication layer
// for testing. Implementations handle the actual wire protocol.
type Transport interface {
	// Connect establishes a connection to the MCP server.
	Connect(ctx context.Context) error
	// Close terminates the connection.
	Close() error
	// Ping sends a health check request and returns an error
	// if the server does not respond.
	Ping(ctx context.Context) error
	// Call invokes an MCP tool by name with the given
	// arguments and returns the raw response.
	Call(
		ctx context.Context,
		tool string,
		args map[string]any,
	) ([]byte, error)
	// ReadResource reads an MCP resource by URI and
	// returns the raw content.
	ReadResource(
		ctx context.Context,
		uri string,
	) ([]byte, error)
	// ListPrompts queries the MCP server for available
	// prompts and returns the raw response.
	ListPrompts(ctx context.Context) ([]byte, error)
}

// ClientConfig holds configuration for the MCP client.
type ClientConfig struct {
	// HealthCheckTimeout is the maximum duration to wait for
	// a health check response.
	HealthCheckTimeout time.Duration
}

// DefaultClientConfig returns a ClientConfig with sensible
// defaults.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		HealthCheckTimeout: 5 * time.Second,
	}
}

// Client communicates with the Gemara MCP server.
type Client struct {
	transport Transport
	config    ClientConfig
	status    ConnectionStatus
	mu        sync.RWMutex
}

// NewClient creates a new MCP client with the given transport
// and configuration.
func NewClient(
	transport Transport,
	config ClientConfig,
) *Client {
	return &Client{
		transport: transport,
		config:    config,
		status:    StatusDisconnected,
	}
}

// Connect establishes a connection to the MCP server.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.transport.Connect(ctx); err != nil {
		return err
	}
	c.status = StatusConnected
	return nil
}

// Close terminates the connection to the MCP server.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status == StatusDisconnected {
		return nil
	}
	err := c.transport.Close()
	c.status = StatusDisconnected
	return err
}

// Status returns the current connection status.
func (c *Client) Status() ConnectionStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// HealthCheck verifies the MCP server is responsive.
func (c *Client) HealthCheck(ctx context.Context) error {
	c.mu.RLock()
	status := c.status
	c.mu.RUnlock()

	if status != StatusConnected {
		return ErrNotConnected
	}

	checkCtx, cancel := context.WithTimeout(
		ctx,
		c.config.HealthCheckTimeout,
	)
	defer cancel()

	if err := c.transport.Ping(checkCtx); err != nil {
		// Mark as disconnected on failure.
		c.mu.Lock()
		c.status = StatusDisconnected
		c.mu.Unlock()
		return ErrHealthCheckFailed
	}
	return nil
}

// GetLexicon reads the gemara://lexicon MCP resource and
// returns the raw response.
func (c *Client) GetLexicon(
	ctx context.Context,
) ([]byte, error) {
	return c.readResource(ctx, "gemara://lexicon")
}

// ValidateArtifact invokes the validate_gemara_artifact MCP
// tool with the given artifact content and schema type.
func (c *Client) ValidateArtifact(
	ctx context.Context,
	artifact string,
	schemaType string,
) ([]byte, error) {
	args := map[string]any{
		"artifact":    artifact,
		"schema_type": schemaType,
	}
	return c.callTool(
		ctx,
		"validate_gemara_artifact",
		args,
	)
}

// GetSchemaDocs reads the gemara://schema/definitions MCP
// resource and returns the raw response.
func (c *Client) GetSchemaDocs(
	ctx context.Context,
) ([]byte, error) {
	return c.readResource(
		ctx, "gemara://schema/definitions",
	)
}

// GetSchemaDocsForVersion reads the
// gemara://schema/definitions resource with a specific
// version parameter.
func (c *Client) GetSchemaDocsForVersion(
	ctx context.Context,
	version string,
) ([]byte, error) {
	uri := "gemara://schema/definitions?version=" +
		version
	return c.readResource(ctx, uri)
}

// MCPPrompt describes an MCP server prompt (wizard).
type MCPPrompt struct {
	// Name is the prompt identifier (e.g., "threat_assessment").
	Name string `json:"name"`
	// Title is the display title.
	Title string `json:"title"`
	// Description explains what the prompt does.
	Description string `json:"description"`
	// Arguments are the required/optional parameters.
	Arguments []MCPPromptArg `json:"arguments"`
}

// MCPPromptArg describes a single prompt argument.
type MCPPromptArg struct {
	// Name is the argument identifier.
	Name string `json:"name"`
	// Title is the display title.
	Title string `json:"title"`
	// Description explains what the argument is for.
	Description string `json:"description"`
	// Required indicates the argument must be provided.
	Required bool `json:"required"`
}

// ListPrompts queries the MCP server for available prompts
// (wizards). Prompts are only available when the server is
// running in artifact mode.
func (c *Client) ListPrompts(
	ctx context.Context,
) ([]MCPPrompt, error) {
	c.mu.RLock()
	status := c.status
	c.mu.RUnlock()

	if status != StatusConnected {
		return nil, ErrNotConnected
	}

	resp, err := c.transport.ListPrompts(ctx)
	if err != nil {
		c.mu.Lock()
		c.status = StatusDisconnected
		c.mu.Unlock()
		return nil, err
	}
	var result struct {
		Prompts []MCPPrompt `json:"prompts"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Prompts, nil
}

// callTool is a helper that checks connection status before
// invoking a tool via the transport.
func (c *Client) callTool(
	ctx context.Context,
	tool string,
	args map[string]any,
) ([]byte, error) {
	c.mu.RLock()
	status := c.status
	c.mu.RUnlock()

	if status != StatusConnected {
		return nil, ErrNotConnected
	}

	resp, err := c.transport.Call(ctx, tool, args)
	if err != nil {
		// Detect disconnection: mark client as disconnected
		// so the session can trigger fallback.
		c.mu.Lock()
		c.status = StatusDisconnected
		c.mu.Unlock()
		return nil, err
	}
	return resp, nil
}

// readResource is a helper that checks connection status
// before reading a resource via the transport.
func (c *Client) readResource(
	ctx context.Context,
	uri string,
) ([]byte, error) {
	c.mu.RLock()
	status := c.status
	c.mu.RUnlock()

	if status != StatusConnected {
		return nil, ErrNotConnected
	}

	resp, err := c.transport.ReadResource(ctx, uri)
	if err != nil {
		// Detect disconnection: mark client as
		// disconnected so the session can trigger
		// fallback.
		c.mu.Lock()
		c.status = StatusDisconnected
		c.mu.Unlock()
		return nil, err
	}
	return resp, nil
}
