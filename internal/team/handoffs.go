// SPDX-License-Identifier: Apache-2.0

package team

import (
	"sort"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
	"github.com/hbraswelrh/gemara-user-journey/internal/tutorials"
)

// HandoffPoint represents a boundary where one team member
// produces artifacts that another team member consumes.
type HandoffPoint struct {
	// ProducerName is the name of the producing member.
	ProducerName string
	// ProducerRole is the role of the producing member.
	ProducerRole string
	// ConsumerName is the name of the consuming member.
	ConsumerName string
	// ConsumerRole is the role of the consuming member.
	ConsumerRole string
	// ProducerLayer is the Gemara layer where artifacts
	// are produced.
	ProducerLayer int
	// ConsumerLayer is the Gemara layer where artifacts
	// are consumed.
	ConsumerLayer int
	// ArtifactTypes lists the artifact types that flow
	// across this handoff boundary.
	ArtifactTypes []string
	// ProducerTutorials lists tutorial titles relevant to
	// the producing role's layer.
	ProducerTutorials []string
	// ConsumerTutorials lists tutorial titles relevant to
	// the consuming role's layer.
	ConsumerTutorials []string
	// Description explains how artifacts flow at this
	// handoff point.
	Description string
}

// ArtifactFlow describes how an artifact type flows between
// Gemara layers.
type ArtifactFlow struct {
	// ArtifactType is the Gemara artifact type name.
	ArtifactType string
	// SourceLayer is the layer that produces the artifact.
	SourceLayer int
	// TargetLayer is the layer that consumes the artifact.
	TargetLayer int
	// Description explains the flow.
	Description string
}

// CollaborationView is the computed view of a team's
// cross-functional interactions within the Gemara model.
type CollaborationView struct {
	// TeamName is the team's display name.
	TeamName string
	// Members are the team members with their layer
	// mappings.
	Members []TeamMember
	// Handoffs are the detected handoff points between
	// team members.
	Handoffs []HandoffPoint
	// CoverageGaps are Gemara layers (1-7) that no team
	// member owns.
	CoverageGaps []int
}

// GenerateView computes the collaboration view for a team
// configuration. It detects handoff points between members
// at layer boundaries, maps artifact flows, identifies
// coverage gaps, and attaches tutorial references.
func GenerateView(
	tc *TeamConfig,
	tuts []tutorials.Tutorial,
) *CollaborationView {
	view := &CollaborationView{
		TeamName: tc.Name,
		Members:  tc.Members,
	}

	view.Handoffs = detectHandoffs(tc, tuts)
	view.CoverageGaps = DetectCoverageGaps(tc)

	return view
}

// DetectCoverageGaps returns Gemara layer numbers (1-7)
// that no team member owns.
func DetectCoverageGaps(tc *TeamConfig) []int {
	covered := make(map[int]bool)
	for _, m := range tc.Members {
		for _, l := range m.Layers {
			covered[l] = true
		}
	}

	var gaps []int
	for layer := consts.LayerGuidance; layer <=
		consts.LayerReporting; layer++ {
		if !covered[layer] {
			gaps = append(gaps, layer)
		}
	}
	return gaps
}

// ArtifactFlows returns the defined layer-to-layer artifact
// flow relationships.
func ArtifactFlows() []ArtifactFlow {
	var flows []ArtifactFlow
	for pair, desc := range consts.ArtifactFlowDescriptions {
		srcLayer := pair[0]
		tgtLayer := pair[1]
		artifacts := consts.LayerArtifacts[srcLayer]
		for _, at := range artifacts {
			flows = append(flows, ArtifactFlow{
				ArtifactType: at,
				SourceLayer:  srcLayer,
				TargetLayer:  tgtLayer,
				Description:  desc,
			})
		}
	}
	// Sort for deterministic output.
	sort.Slice(flows, func(i, j int) bool {
		if flows[i].SourceLayer !=
			flows[j].SourceLayer {
			return flows[i].SourceLayer <
				flows[j].SourceLayer
		}
		return flows[i].TargetLayer <
			flows[j].TargetLayer
	})
	return flows
}

// detectHandoffs finds handoff points between team members.
// A handoff occurs when one member owns a layer that
// produces artifacts consumed by a layer owned by a
// different member.
func detectHandoffs(
	tc *TeamConfig,
	tuts []tutorials.Tutorial,
) []HandoffPoint {
	// Build a tutorial index by layer.
	tutsByLayer := make(map[int][]string)
	for _, t := range tuts {
		tutsByLayer[t.Layer] = append(
			tutsByLayer[t.Layer], t.Title,
		)
	}

	// Build member-to-layers index.
	memberLayers := make(map[string]map[int]bool)
	for _, m := range tc.Members {
		layers := make(map[int]bool)
		for _, l := range m.Layers {
			layers[l] = true
		}
		memberLayers[m.Name] = layers
	}

	var handoffs []HandoffPoint

	// Check each pair of members for layer boundary
	// handoffs.
	for i := 0; i < len(tc.Members); i++ {
		for j := 0; j < len(tc.Members); j++ {
			if i == j {
				continue
			}
			producer := tc.Members[i]
			consumer := tc.Members[j]

			for _, pLayer := range producer.Layers {
				for _, cLayer := range consumer.Layers {
					if pLayer >= cLayer {
						continue
					}
					// Check if there's a defined flow
					// between these layers.
					desc, exists :=
						consts.ArtifactFlowDescriptions[[2]int{pLayer, cLayer}]
					if !exists {
						continue
					}

					artifacts :=
						consts.LayerArtifacts[pLayer]
					if len(artifacts) == 0 {
						continue
					}

					hp := HandoffPoint{
						ProducerName:  producer.Name,
						ProducerRole:  producer.RoleName,
						ConsumerName:  consumer.Name,
						ConsumerRole:  consumer.RoleName,
						ProducerLayer: pLayer,
						ConsumerLayer: cLayer,
						ArtifactTypes: artifacts,
						Description:   desc,
					}

					hp.ProducerTutorials =
						tutsByLayer[pLayer]
					hp.ConsumerTutorials =
						tutsByLayer[cLayer]

					handoffs = append(handoffs, hp)
				}
			}
		}
	}

	// Sort for deterministic output.
	sort.Slice(handoffs, func(i, j int) bool {
		if handoffs[i].ProducerLayer !=
			handoffs[j].ProducerLayer {
			return handoffs[i].ProducerLayer <
				handoffs[j].ProducerLayer
		}
		if handoffs[i].ConsumerLayer !=
			handoffs[j].ConsumerLayer {
			return handoffs[i].ConsumerLayer <
				handoffs[j].ConsumerLayer
		}
		if handoffs[i].ProducerName !=
			handoffs[j].ProducerName {
			return handoffs[i].ProducerName <
				handoffs[j].ProducerName
		}
		return handoffs[i].ConsumerName <
			handoffs[j].ConsumerName
	})

	return handoffs
}
