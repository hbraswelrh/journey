// SPDX-License-Identifier: Apache-2.0

// Package cli provides the command-line interface for Gemara User Journey,
// including the first-launch setup flow, role discovery, and
// guided authoring commands.
package cli

import (
	"fmt"
	"io"
)

// CLI holds shared state for the command-line interface.
type CLI struct {
	// Out is the writer for user-facing output (stdout).
	Out io.Writer
	// Err is the writer for error output (stderr).
	Err io.Writer
}

// NewCLI creates a new CLI instance with the given output
// writers.
func NewCLI(out, errOut io.Writer) *CLI {
	return &CLI{Out: out, Err: errOut}
}

// Print writes a formatted message to the user-facing output.
func (c *CLI) Print(format string, args ...any) {
	fmt.Fprintf(c.Out, format, args...)
}

// PrintErr writes a formatted message to the error output.
func (c *CLI) PrintErr(format string, args ...any) {
	fmt.Fprintf(c.Err, format, args...)
}
