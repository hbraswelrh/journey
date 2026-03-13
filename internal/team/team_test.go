// SPDX-License-Identifier: Apache-2.0

package team_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hbraswelrh/pacman/internal/team"
)

func TestNewTeamConfig(t *testing.T) {
	tc := team.NewTeamConfig("Security Team")

	if tc.Name != "Security Team" {
		t.Errorf(
			"Name: got %s, want Security Team",
			tc.Name,
		)
	}
	if len(tc.Members) != 0 {
		t.Errorf(
			"Members: got %d, want 0",
			len(tc.Members),
		)
	}
}

func TestAddMember(t *testing.T) {
	tc := team.NewTeamConfig("Test Team")

	member := team.TeamMember{
		Name:     "Alice",
		RoleName: "Security Engineer",
		Layers:   []int{1, 2},
		Keywords: []string{"CI/CD", "SDLC"},
	}

	err := tc.AddMember(member)
	if err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	if len(tc.Members) != 1 {
		t.Fatalf("Members: got %d, want 1",
			len(tc.Members))
	}
	if tc.Members[0].Name != "Alice" {
		t.Errorf(
			"Name: got %s, want Alice",
			tc.Members[0].Name,
		)
	}
	if tc.Members[0].RoleName != "Security Engineer" {
		t.Errorf(
			"RoleName: got %s, want Security Engineer",
			tc.Members[0].RoleName,
		)
	}
	if len(tc.Members[0].Layers) != 2 {
		t.Errorf(
			"Layers: got %d, want 2",
			len(tc.Members[0].Layers),
		)
	}
}

func TestAddMemberDuplicate(t *testing.T) {
	tc := team.NewTeamConfig("Test Team")

	member := team.TeamMember{
		Name:     "Alice",
		RoleName: "Security Engineer",
		Layers:   []int{1, 2},
	}

	if err := tc.AddMember(member); err != nil {
		t.Fatalf("first AddMember: %v", err)
	}

	err := tc.AddMember(member)
	if err == nil {
		t.Fatal("expected error for duplicate member")
	}
}

func TestRemoveMember(t *testing.T) {
	tc := team.NewTeamConfig("Test Team")

	_ = tc.AddMember(team.TeamMember{
		Name:     "Alice",
		RoleName: "Security Engineer",
		Layers:   []int{1, 2},
	})

	if !tc.RemoveMember("Alice") {
		t.Fatal("expected RemoveMember to return true")
	}
	if len(tc.Members) != 0 {
		t.Errorf(
			"Members: got %d, want 0",
			len(tc.Members),
		)
	}

	if tc.RemoveMember("Nonexistent") {
		t.Fatal(
			"expected RemoveMember to return false " +
				"for nonexistent member",
		)
	}
}

func TestSaveAndLoadTeam(t *testing.T) {
	dir := t.TempDir()

	tc := team.NewTeamConfig("Security Team")
	_ = tc.AddMember(team.TeamMember{
		Name:     "Alice",
		RoleName: "Security Engineer",
		Layers:   []int{1, 2},
		Keywords: []string{"CI/CD"},
	})
	_ = tc.AddMember(team.TeamMember{
		Name:     "Bob",
		RoleName: "Compliance Officer",
		Layers:   []int{3, 5},
		Keywords: []string{"audit"},
	})

	if err := team.SaveTeam(dir, tc); err != nil {
		t.Fatalf("SaveTeam: %v", err)
	}

	files, _ := os.ReadDir(dir)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	path := filepath.Join(dir, files[0].Name())
	loaded, err := team.LoadTeam(path)
	if err != nil {
		t.Fatalf("LoadTeam: %v", err)
	}

	if loaded.Name != "Security Team" {
		t.Errorf(
			"Name: got %s, want Security Team",
			loaded.Name,
		)
	}
	if len(loaded.Members) != 2 {
		t.Errorf(
			"Members: got %d, want 2",
			len(loaded.Members),
		)
	}
}

func TestLoadTeamNonexistent(t *testing.T) {
	_, err := team.LoadTeam("/nonexistent/team.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestListTeams(t *testing.T) {
	dir := t.TempDir()

	tc1 := team.NewTeamConfig("Team Alpha")
	_ = tc1.AddMember(team.TeamMember{
		Name:     "Alice",
		RoleName: "Developer",
		Layers:   []int{2, 4},
	})

	tc2 := team.NewTeamConfig("Team Beta")
	_ = tc2.AddMember(team.TeamMember{
		Name:     "Bob",
		RoleName: "Auditor",
		Layers:   []int{5, 3},
	})

	_ = team.SaveTeam(dir, tc1)
	_ = team.SaveTeam(dir, tc2)

	teams, err := team.ListTeams(dir)
	if err != nil {
		t.Fatalf("ListTeams: %v", err)
	}
	if len(teams) != 2 {
		t.Errorf("Teams: got %d, want 2", len(teams))
	}
}

func TestListTeamsMissingDir(t *testing.T) {
	teams, err := team.ListTeams("/nonexistent/dir")
	if err != nil {
		t.Fatalf("ListTeams should not error: %v", err)
	}
	if len(teams) != 0 {
		t.Errorf("Teams: got %d, want 0", len(teams))
	}
}
