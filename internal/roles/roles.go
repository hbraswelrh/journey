// SPDX-License-Identifier: Apache-2.0

// Package roles implements role identification, activity
// probing, keyword extraction, and layer mapping for the
// Gemara User Journey role-based tutorial engine.
package roles

import (
	"strings"

	"github.com/hbraswelrh/journey/internal/consts"
)

// RoleSource indicates whether a role is predefined or custom.
type RoleSource int

const (
	// SourcePredefined is a built-in role.
	SourcePredefined RoleSource = iota
	// SourceCustom is a user-defined role.
	SourceCustom
)

// Role represents a job function that serves as the starting
// point for tutorial routing.
type Role struct {
	// Name is the display name of the role.
	Name string
	// Description explains the role's typical
	// responsibilities.
	Description string
	// DefaultKeywords are the activity keywords associated
	// with this role by default.
	DefaultKeywords []string
	// DefaultLayers are the Gemara layers this role
	// typically interacts with.
	DefaultLayers []int
	// Source indicates whether the role is predefined or
	// custom.
	Source RoleSource
}

// MatchType describes the quality of a role match.
type MatchType int

const (
	// MatchNone means no match was found.
	MatchNone MatchType = iota
	// MatchPartial means the input partially overlaps with
	// a predefined role (e.g., "Product Security Engineer"
	// contains "Security Engineer").
	MatchPartial
	// MatchExact means the input exactly matches a
	// predefined role name.
	MatchExact
)

// MatchResult holds the outcome of matching user input against
// predefined roles.
type MatchResult struct {
	// Role is the matched role, if any.
	Role *Role
	// Type is the quality of the match.
	Type MatchType
	// OverlappingKeywords are predefined keywords found in
	// the user input.
	OverlappingKeywords []string
	// Confidence is a 0-1 score indicating match quality.
	Confidence float64
}

// PredefinedRoles returns the minimum required list of roles
// per FR-002. This list is extensible through configuration.
func PredefinedRoles() []Role {
	return []Role{
		{
			Name: consts.RoleSecurityEngineer,
			Description: "Identifies threats, authors " +
				"controls, and reviews secure " +
				"architecture.",
			DefaultKeywords: []string{
				"threat modeling",
				"penetration testing",
				"secure architecture review",
				"CI/CD",
				"dependency management",
				"SDLC",
			},
			DefaultLayers: []int{
				consts.LayerThreatsControls,
				consts.LayerGuidance,
			},
			Source: SourcePredefined,
		},
		{
			Name: consts.RoleComplianceOfficer,
			Description: "Manages regulatory alignment, " +
				"evidence collection, and audit " +
				"preparation.",
			DefaultKeywords: []string{
				"evidence collection",
				"audit interviews",
				"assessment plans",
				"scope definition",
				"adherence",
			},
			DefaultLayers: []int{
				consts.LayerRiskPolicy,
				consts.LayerGuidance,
				consts.LayerEvaluation,
			},
			Source: SourcePredefined,
		},
		{
			Name: consts.RoleCISO,
			Description: "Sets security strategy, defines " +
				"risk appetite, and oversees policy.",
			DefaultKeywords: []string{
				"risk appetite",
				"create policy",
				"scope definition",
				"non-compliance handling",
			},
			DefaultLayers: []int{
				consts.LayerRiskPolicy,
				consts.LayerGuidance,
			},
			Source: SourcePredefined,
		},
		{
			Name: consts.RoleDeveloper,
			Description: "Writes code, manages CI/CD " +
				"pipelines, and integrates security " +
				"controls.",
			DefaultKeywords: []string{
				"CI/CD",
				"dependency management",
				"upstream open-source",
				"SDLC",
			},
			DefaultLayers: []int{
				consts.LayerThreatsControls,
				consts.LayerSensitiveActivity,
			},
			Source: SourcePredefined,
		},
		{
			Name: consts.RolePlatformEngineer,
			Description: "Builds and operates " +
				"infrastructure with security " +
				"controls embedded.",
			DefaultKeywords: []string{
				"CI/CD",
				"pipeline security",
				"dependency management",
				"upstream open-source",
			},
			DefaultLayers: []int{
				consts.LayerThreatsControls,
				consts.LayerSensitiveActivity,
			},
			Source: SourcePredefined,
		},
		{
			Name: consts.RolePolicyAuthor,
			Description: "Drafts organizational policies " +
				"and defines adherence timelines.",
			DefaultKeywords: []string{
				"create policy",
				"timeline for adherence",
				"adherence requirements",
				"scope definition",
			},
			DefaultLayers: []int{
				consts.LayerRiskPolicy,
			},
			Source: SourcePredefined,
		},
		{
			Name: consts.RoleAuditor,
			Description: "Evaluates compliance, conducts " +
				"assessments, and reviews evidence.",
			DefaultKeywords: []string{
				"audit interviews",
				"assessment plans",
				"evidence collection",
				"evaluation",
			},
			DefaultLayers: []int{
				consts.LayerEvaluation,
				consts.LayerRiskPolicy,
			},
			Source: SourcePredefined,
		},
	}
}

// MatchRole compares free-text input against predefined roles
// and returns the best match. Partial matches (e.g., "Product
// Security Engineer" contains "Security Engineer") are
// identified but NOT assumed to be exact — the caller should
// proceed to activity probing for refinement (FR-023).
func MatchRole(input string) MatchResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return MatchResult{Type: MatchNone}
	}

	lower := strings.ToLower(input)
	predefined := PredefinedRoles()

	// Check for exact match first.
	for i := range predefined {
		if strings.EqualFold(input, predefined[i].Name) {
			return MatchResult{
				Role:       &predefined[i],
				Type:       MatchExact,
				Confidence: 1.0,
			}
		}
	}

	// Check for partial match.
	var bestMatch *Role
	var bestOverlap []string
	bestConfidence := 0.0

	for i := range predefined {
		roleLower := strings.ToLower(predefined[i].Name)

		// Check if the predefined role name appears in
		// the input or vice versa.
		if strings.Contains(lower, roleLower) ||
			strings.Contains(roleLower, lower) {
			overlap := findOverlappingKeywords(
				lower, predefined[i].DefaultKeywords,
			)
			// Confidence based on name overlap length
			// relative to input length.
			conf := float64(len(roleLower)) /
				float64(len(lower))
			if conf > 1.0 {
				conf = 1.0 / conf
			}
			if conf > bestConfidence {
				bestMatch = &predefined[i]
				bestOverlap = overlap
				bestConfidence = conf
			}
		}

		// Also check keyword overlap for non-name
		// matches.
		overlap := findOverlappingKeywords(
			lower, predefined[i].DefaultKeywords,
		)
		if len(overlap) > len(bestOverlap) {
			kwConf := float64(len(overlap)) /
				float64(
					len(predefined[i].DefaultKeywords),
				)
			if kwConf > bestConfidence {
				bestMatch = &predefined[i]
				bestOverlap = overlap
				bestConfidence = kwConf * 0.8
			}
		}
	}

	if bestMatch != nil {
		return MatchResult{
			Role:                bestMatch,
			Type:                MatchPartial,
			OverlappingKeywords: bestOverlap,
			Confidence:          bestConfidence,
		}
	}

	return MatchResult{Type: MatchNone}
}

// findOverlappingKeywords returns predefined keywords that
// appear in the input text.
func findOverlappingKeywords(
	lowerInput string,
	keywords []string,
) []string {
	var found []string
	for _, kw := range keywords {
		if strings.Contains(
			lowerInput, strings.ToLower(kw),
		) {
			found = append(found, kw)
		}
	}
	return found
}
