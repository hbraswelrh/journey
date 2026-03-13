// SPDX-License-Identifier: Apache-2.0

// Package team manages cross-functional team configurations
// and collaboration views, mapping team members' roles to
// Gemara layers and identifying handoff points.
package team

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// TeamMember represents a single team member with their
// resolved role and Gemara layer mappings.
type TeamMember struct {
	// Name is the team member's display name.
	Name string `yaml:"name"`
	// RoleName is the role associated with this member.
	RoleName string `yaml:"role_name"`
	// Layers are the Gemara layer numbers this member
	// interacts with.
	Layers []int `yaml:"layers"`
	// Keywords are the activity keywords associated with
	// this member's role.
	Keywords []string `yaml:"keywords,omitempty"`
}

// TeamConfig holds the configuration for a cross-functional
// team.
type TeamConfig struct {
	// Name is the team's display name.
	Name string `yaml:"name"`
	// Members are the team's members.
	Members []TeamMember `yaml:"members"`
}

// NewTeamConfig creates a new team configuration with the
// given name and an empty member list.
func NewTeamConfig(name string) *TeamConfig {
	return &TeamConfig{
		Name:    name,
		Members: []TeamMember{},
	}
}

// AddMember adds a team member. Returns an error if a member
// with the same name already exists.
func (tc *TeamConfig) AddMember(
	member TeamMember,
) error {
	for _, m := range tc.Members {
		if m.Name == member.Name {
			return fmt.Errorf(
				"member %q already exists in team %q",
				member.Name, tc.Name,
			)
		}
	}
	tc.Members = append(tc.Members, member)
	return nil
}

// RemoveMember removes a team member by name. Returns true
// if the member was found and removed, false otherwise.
func (tc *TeamConfig) RemoveMember(name string) bool {
	for i, m := range tc.Members {
		if m.Name == name {
			tc.Members = append(
				tc.Members[:i], tc.Members[i+1:]...,
			)
			return true
		}
	}
	return false
}

// SaveTeam writes a team configuration as a YAML file in
// the given directory. The filename is derived from the team
// name.
func SaveTeam(dir string, tc *TeamConfig) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf(
			"create team directory %s: %w", dir, err,
		)
	}

	data, err := yaml.Marshal(tc)
	if err != nil {
		return fmt.Errorf("marshal team config: %w", err)
	}

	filename := sanitizeFilename(tc.Name) + ".yaml"
	path := filepath.Join(dir, filename)

	if err := os.WriteFile(
		path, data, 0o644,
	); err != nil {
		return fmt.Errorf(
			"write team config %s: %w", path, err,
		)
	}

	return nil
}

// LoadTeam reads a single team configuration from a YAML
// file.
func LoadTeam(path string) (*TeamConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"read team config %s: %w", path, err,
		)
	}

	var tc TeamConfig
	if err := yaml.Unmarshal(data, &tc); err != nil {
		return nil, fmt.Errorf(
			"parse team config %s: %w", path, err,
		)
	}

	return &tc, nil
}

// ListTeams reads all YAML team config files from a
// directory and returns the parsed configurations.
func ListTeams(dir string) ([]TeamConfig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf(
			"read team directory %s: %w", dir, err,
		)
	}

	var teams []TeamConfig
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		tc, err := LoadTeam(path)
		if err != nil {
			// Skip unparseable files.
			continue
		}
		teams = append(teams, *tc)
	}

	return teams, nil
}

// sanitizeFilename converts a team name to a safe filename.
func sanitizeFilename(name string) string {
	safe := strings.ToLower(name)
	safe = strings.ReplaceAll(safe, " ", "-")
	safe = strings.ReplaceAll(safe, "/", "-")
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
