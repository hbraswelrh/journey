// SPDX-License-Identifier: Apache-2.0

package cli

// ExportUserHomeDir returns the current userHomeDir function
// for saving/restoring in tests.
func ExportUserHomeDir() func() (string, error) {
	return userHomeDir
}

// SetUserHomeDir replaces the userHomeDir function for
// testing.
func SetUserHomeDir(fn func() (string, error)) {
	userHomeDir = fn
}
