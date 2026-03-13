// SPDX-License-Identifier: Apache-2.0

package roles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// RoleProfile is a persistable role + activity configuration
// that can be saved for reuse across sessions and assigned to
// team members.
type RoleProfile struct {
	// Name is the display name of the role.
	Name string `yaml:"name"`
	// ActivityKeywords are the resolved keywords for this
	// profile.
	ActivityKeywords []string `yaml:"activity_keywords"`
	// LayerMappings are the resolved Gemara layer numbers.
	LayerMappings []int `yaml:"layer_mappings"`
	// Description is an optional free-text description.
	Description string `yaml:"description,omitempty"`
	// CreatedAt is the timestamp when the profile was
	// saved.
	CreatedAt time.Time `yaml:"created_at"`
}

// SaveProfile writes a role profile as a YAML file in the
// given directory. The filename is derived from the profile
// name.
func SaveProfile(
	dir string,
	profile *RoleProfile,
) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf(
			"create profile directory %s: %w",
			dir, err,
		)
	}

	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = time.Now()
	}

	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}

	filename := sanitizeFilename(profile.Name) + ".yaml"
	path := filepath.Join(dir, filename)

	if err := os.WriteFile(
		path, data, 0o644,
	); err != nil {
		return fmt.Errorf(
			"write profile %s: %w", path, err,
		)
	}

	return nil
}

// LoadProfile reads a single role profile from a YAML file.
func LoadProfile(path string) (*RoleProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"read profile %s: %w", path, err,
		)
	}

	var profile RoleProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf(
			"parse profile %s: %w", path, err,
		)
	}

	return &profile, nil
}

// ListProfiles reads all YAML profile files from a directory
// and returns the parsed profiles.
func ListProfiles(dir string) ([]RoleProfile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf(
			"read profile directory %s: %w",
			dir, err,
		)
	}

	var profiles []RoleProfile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		profile, err := LoadProfile(path)
		if err != nil {
			// Skip unparseable files.
			continue
		}
		profiles = append(profiles, *profile)
	}

	return profiles, nil
}

// MergeWithPredefined combines predefined roles with custom
// profiles into a unified selection list. Custom profiles are
// appended after predefined roles.
func MergeWithPredefined(
	predefined []Role,
	custom []RoleProfile,
) []Role {
	merged := make(
		[]Role, len(predefined), len(predefined)+len(custom),
	)
	copy(merged, predefined)

	for _, cp := range custom {
		merged = append(merged, Role{
			Name:            cp.Name,
			Description:     cp.Description,
			DefaultKeywords: cp.ActivityKeywords,
			DefaultLayers:   cp.LayerMappings,
			Source:          SourceCustom,
		})
	}

	return merged
}

// ProfileFromActivityProfile converts an ActivityProfile to
// a persistable RoleProfile.
func ProfileFromActivityProfile(
	ap *ActivityProfile,
) *RoleProfile {
	name := ""
	desc := ""
	if ap.Role != nil {
		name = ap.Role.Name
		desc = ap.Role.Description
	}

	return &RoleProfile{
		Name:             name,
		ActivityKeywords: ap.ExtractedKeywords,
		LayerMappings:    ap.UniqueLayerNumbers(),
		Description:      desc,
		CreatedAt:        time.Now(),
	}
}

// sanitizeFilename converts a role name to a safe filename.
func sanitizeFilename(name string) string {
	safe := strings.ToLower(name)
	safe = strings.ReplaceAll(safe, " ", "-")
	safe = strings.ReplaceAll(safe, "/", "-")
	// Remove any remaining unsafe characters.
	var result strings.Builder
	for _, r := range safe {
		if (r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
