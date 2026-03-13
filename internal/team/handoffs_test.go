// SPDX-License-Identifier: Apache-2.0

package team_test

import (
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/team"
	"github.com/hbraswelrh/pacman/internal/tutorials"
)

// testTutorials returns a minimal tutorial set for handoff
// testing.
func testTutorials() []tutorials.Tutorial {
	return []tutorials.Tutorial{
		{
			Title:    "Guidance Catalog Guide",
			FilePath: "guidance-catalog-guide.md",
			Layer:    consts.LayerGuidance,
		},
		{
			Title:    "Threat Assessment Guide",
			FilePath: "threat-assessment-guide.md",
			Layer:    consts.LayerThreatsControls,
		},
		{
			Title:    "Control Catalog Guide",
			FilePath: "control-catalog-guide.md",
			Layer:    consts.LayerThreatsControls,
		},
		{
			Title:    "Policy Guide",
			FilePath: "policy-guide.md",
			Layer:    consts.LayerRiskPolicy,
		},
	}
}

func TestGenerateViewThreeRoles(t *testing.T) {
	tc := team.NewTeamConfig("GRC Team")
	_ = tc.AddMember(team.TeamMember{
		Name:     "Alice",
		RoleName: consts.RoleSecurityEngineer,
		Layers: []int{
			consts.LayerGuidance,
			consts.LayerThreatsControls,
		},
	})
	_ = tc.AddMember(team.TeamMember{
		Name:     "Bob",
		RoleName: consts.RoleComplianceOfficer,
		Layers: []int{
			consts.LayerRiskPolicy,
			consts.LayerEvaluation,
		},
	})
	_ = tc.AddMember(team.TeamMember{
		Name:     "Carol",
		RoleName: consts.RoleDeveloper,
		Layers: []int{
			consts.LayerSensitiveActivity,
			consts.LayerEvaluation,
		},
	})

	view := team.GenerateView(tc, testTutorials())

	if view.TeamName != "GRC Team" {
		t.Errorf(
			"TeamName: got %s, want GRC Team",
			view.TeamName,
		)
	}

	if len(view.Members) != 3 {
		t.Errorf(
			"Members: got %d, want 3",
			len(view.Members),
		)
	}

	// Expect handoffs: Alice(L2) -> Bob(L3),
	// Bob(L3) -> Carol(L4).
	if len(view.Handoffs) < 2 {
		t.Fatalf(
			"Handoffs: got %d, want at least 2",
			len(view.Handoffs),
		)
	}

	foundL2L3 := false
	foundL3L4 := false
	for _, hp := range view.Handoffs {
		if hp.ProducerLayer == consts.LayerThreatsControls &&
			hp.ConsumerLayer == consts.LayerRiskPolicy {
			foundL2L3 = true
		}
		if hp.ProducerLayer == consts.LayerRiskPolicy &&
			hp.ConsumerLayer ==
				consts.LayerSensitiveActivity {
			foundL3L4 = true
		}
	}
	if !foundL2L3 {
		t.Error("expected handoff L2 -> L3")
	}
	if !foundL3L4 {
		t.Error("expected handoff L3 -> L4")
	}
}

func TestHandoffArtifactTypes(t *testing.T) {
	tc := team.NewTeamConfig("Test Team")
	_ = tc.AddMember(team.TeamMember{
		Name:     "Alice",
		RoleName: consts.RoleSecurityEngineer,
		Layers:   []int{consts.LayerThreatsControls},
	})
	_ = tc.AddMember(team.TeamMember{
		Name:     "Bob",
		RoleName: consts.RoleComplianceOfficer,
		Layers:   []int{consts.LayerRiskPolicy},
	})

	view := team.GenerateView(tc, testTutorials())

	if len(view.Handoffs) == 0 {
		t.Fatal("expected at least one handoff")
	}

	hp := view.Handoffs[0]
	if len(hp.ArtifactTypes) == 0 {
		t.Fatal("expected artifact types on handoff")
	}

	// L2 produces ThreatCatalog and ControlCatalog.
	hasControl := false
	hasThreat := false
	for _, at := range hp.ArtifactTypes {
		if at == consts.ArtifactControlCatalog {
			hasControl = true
		}
		if at == consts.ArtifactThreatCatalog {
			hasThreat = true
		}
	}
	if !hasControl {
		t.Error(
			"expected ControlCatalog in handoff artifacts",
		)
	}
	if !hasThreat {
		t.Error(
			"expected ThreatCatalog in handoff artifacts",
		)
	}
}

func TestCoverageGaps(t *testing.T) {
	tc := team.NewTeamConfig("Test Team")
	_ = tc.AddMember(team.TeamMember{
		Name:   "Alice",
		Layers: []int{1, 2},
	})
	_ = tc.AddMember(team.TeamMember{
		Name:   "Bob",
		Layers: []int{3, 5},
	})
	_ = tc.AddMember(team.TeamMember{
		Name:   "Carol",
		Layers: []int{4, 5},
	})

	gaps := team.DetectCoverageGaps(tc)

	// Layers 6 and 7 should be uncovered.
	if len(gaps) != 2 {
		t.Fatalf("Gaps: got %d, want 2", len(gaps))
	}
	if gaps[0] != consts.LayerDataCollection {
		t.Errorf(
			"Gap[0]: got %d, want %d",
			gaps[0], consts.LayerDataCollection,
		)
	}
	if gaps[1] != consts.LayerReporting {
		t.Errorf(
			"Gap[1]: got %d, want %d",
			gaps[1], consts.LayerReporting,
		)
	}
}

func TestAddMemberUpdatesHandoffs(t *testing.T) {
	tc := team.NewTeamConfig("Test Team")
	_ = tc.AddMember(team.TeamMember{
		Name:   "Alice",
		Layers: []int{1, 2},
	})
	_ = tc.AddMember(team.TeamMember{
		Name:   "Bob",
		Layers: []int{3},
	})

	view1 := team.GenerateView(tc, testTutorials())
	count1 := len(view1.Handoffs)

	_ = tc.AddMember(team.TeamMember{
		Name:     "Carol",
		RoleName: consts.RoleAuditor,
		Layers:   []int{5, 3},
	})

	view2 := team.GenerateView(tc, testTutorials())
	if len(view2.Handoffs) <= count1 {
		t.Errorf(
			"expected more handoffs after adding member:"+
				" got %d, had %d",
			len(view2.Handoffs), count1,
		)
	}
}

func TestSameLayerNoHandoff(t *testing.T) {
	tc := team.NewTeamConfig("Test Team")
	_ = tc.AddMember(team.TeamMember{
		Name:   "Alice",
		Layers: []int{5},
	})
	_ = tc.AddMember(team.TeamMember{
		Name:   "Bob",
		Layers: []int{5},
	})

	view := team.GenerateView(tc, testTutorials())

	// Two members at the same layer should not produce a
	// handoff between them for that layer.
	for _, hp := range view.Handoffs {
		if hp.ProducerLayer == hp.ConsumerLayer {
			t.Errorf(
				"unexpected same-layer handoff at L%d",
				hp.ProducerLayer,
			)
		}
	}
}

func TestSingleMemberNoHandoffs(t *testing.T) {
	tc := team.NewTeamConfig("Solo Team")
	_ = tc.AddMember(team.TeamMember{
		Name:   "Alice",
		Layers: []int{1, 2, 3},
	})

	view := team.GenerateView(tc, testTutorials())

	if len(view.Handoffs) != 0 {
		t.Errorf(
			"Handoffs: got %d, want 0",
			len(view.Handoffs),
		)
	}
}

func TestHandoffTutorialReferences(t *testing.T) {
	tc := team.NewTeamConfig("Test Team")
	_ = tc.AddMember(team.TeamMember{
		Name:   "Alice",
		Layers: []int{consts.LayerThreatsControls},
	})
	_ = tc.AddMember(team.TeamMember{
		Name:   "Bob",
		Layers: []int{consts.LayerRiskPolicy},
	})

	view := team.GenerateView(tc, testTutorials())

	if len(view.Handoffs) == 0 {
		t.Fatal("expected at least one handoff")
	}

	hp := view.Handoffs[0]
	if len(hp.ProducerTutorials) == 0 {
		t.Error("expected producer tutorial references")
	}
	if len(hp.ConsumerTutorials) == 0 {
		t.Error("expected consumer tutorial references")
	}
}
