// SPDX-License-Identifier: Apache-2.0

package mcp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hbraswelrh/pacman/internal/mcp"
)

// mockTransport implements mcp.Transport for testing.
type mockTransport struct {
	connectErr error
	closeErr   error
	pingErr    error
	callResp   []byte
	callErr    error
}

func (m *mockTransport) Connect(
	_ context.Context,
) error {
	return m.connectErr
}

func (m *mockTransport) Close() error {
	return m.closeErr
}

func (m *mockTransport) Ping(_ context.Context) error {
	return m.pingErr
}

func (m *mockTransport) Call(
	_ context.Context,
	_ string,
	_ map[string]any,
) ([]byte, error) {
	return m.callResp, m.callErr
}

func TestClient_HealthCheckSucceeds(t *testing.T) {
	transport := &mockTransport{}
	client := mcp.NewClient(
		transport,
		mcp.DefaultClientConfig(),
	)

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("unexpected connect error: %v", err)
	}

	if err := client.HealthCheck(ctx); err != nil {
		t.Fatalf("unexpected health check error: %v", err)
	}

	if client.Status() != mcp.StatusConnected {
		t.Fatal("expected StatusConnected after health check")
	}
}

func TestClient_HealthCheckFailsWithTimeout(t *testing.T) {
	transport := &mockTransport{
		pingErr: context.DeadlineExceeded,
	}
	cfg := mcp.ClientConfig{
		HealthCheckTimeout: 50 * time.Millisecond,
	}
	client := mcp.NewClient(transport, cfg)

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("unexpected connect error: %v", err)
	}

	err := client.HealthCheck(ctx)
	if !errors.Is(err, mcp.ErrHealthCheckFailed) {
		t.Fatalf(
			"expected ErrHealthCheckFailed, got: %v",
			err,
		)
	}

	if client.Status() != mcp.StatusDisconnected {
		t.Fatal(
			"expected StatusDisconnected after failed " +
				"health check",
		)
	}
}

func TestClient_MidSessionDisconnection(t *testing.T) {
	transport := &mockTransport{
		callErr: errors.New("connection reset"),
	}
	client := mcp.NewClient(
		transport,
		mcp.DefaultClientConfig(),
	)

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("unexpected connect error: %v", err)
	}

	if client.Status() != mcp.StatusConnected {
		t.Fatal("expected StatusConnected after connect")
	}

	// Attempt a tool call that fails due to disconnection.
	_, err := client.GetLexicon(ctx)
	if err == nil {
		t.Fatal("expected error from disconnected call")
	}

	// Client should detect disconnection and update status.
	if client.Status() != mcp.StatusDisconnected {
		t.Fatal(
			"expected StatusDisconnected after failed " +
				"tool call",
		)
	}
}

func TestClient_OperationWhenNotConnected(t *testing.T) {
	transport := &mockTransport{}
	client := mcp.NewClient(
		transport,
		mcp.DefaultClientConfig(),
	)

	ctx := context.Background()

	// Health check without connecting should fail.
	err := client.HealthCheck(ctx)
	if !errors.Is(err, mcp.ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected, got: %v", err)
	}

	// Tool call without connecting should fail.
	_, err = client.GetLexicon(ctx)
	if !errors.Is(err, mcp.ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected, got: %v", err)
	}
}

func TestClient_GetLexiconSuccess(t *testing.T) {
	expected := []byte(`{"terms": []}`)
	transport := &mockTransport{callResp: expected}
	client := mcp.NewClient(
		transport,
		mcp.DefaultClientConfig(),
	)

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("unexpected connect error: %v", err)
	}

	resp, err := client.GetLexicon(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != string(expected) {
		t.Fatalf(
			"expected %q, got %q",
			string(expected),
			string(resp),
		)
	}
}

func TestClient_ValidateArtifactSuccess(t *testing.T) {
	expected := []byte(`{"valid": true}`)
	transport := &mockTransport{callResp: expected}
	client := mcp.NewClient(
		transport,
		mcp.DefaultClientConfig(),
	)

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("unexpected connect error: %v", err)
	}

	resp, err := client.ValidateArtifact(
		ctx,
		"content: test",
		"#GuidanceCatalog",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != string(expected) {
		t.Fatalf(
			"expected %q, got %q",
			string(expected),
			string(resp),
		)
	}
}
