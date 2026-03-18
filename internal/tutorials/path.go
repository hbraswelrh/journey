// SPDX-License-Identifier: Apache-2.0

package tutorials

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hbraswelrh/pacman/internal/roles"
)

// LearningPath is an ordered sequence of tutorial references
// tailored to a specific role and activity profile.
type LearningPath struct {
	// TargetRole is the role name this path was built for.
	TargetRole string
	// Steps are the ordered path steps.
	Steps []PathStep
	// CompletedSteps tracks which steps have been
	// completed by index.
	CompletedSteps map[int]bool
}

// PathStep is a single item within a learning path.
type PathStep struct {
	// Tutorial is the source tutorial.
	Tutorial Tutorial
	// Layer is the Gemara layer this step targets.
	Layer int
	// WhyAnnotation explains why this tutorial matters
	// for the user's role and activities.
	WhyAnnotation string
	// HowAnnotation explains how the user will apply
	// the concepts.
	HowAnnotation string
	// WhatAnnotation summarizes what the tutorial covers.
	WhatAnnotation string
	// Prerequisites are indices of steps that should be
	// completed before this one.
	Prerequisites []int
	// VersionMismatch is set when the tutorial's schema
	// version differs from the selected version.
	VersionMismatch *VersionMismatch
	// SectionRelevance maps section headings to relevance
	// scores based on the user's activity keywords.
	// Sections with score > 0 match the user's stated
	// activities.
	SectionRelevance map[string]int
	// PrimarySections lists section headings that are
	// most relevant to the user's stated activities,
	// ordered by relevance score descending.
	PrimarySections []string
}

// StepNavInfo provides navigation context for a path step,
// including prerequisite warnings for non-linear navigation.
type StepNavInfo struct {
	// StepIndex is the position of this step in the path.
	StepIndex int
	// IsCompleted is true if this step has been marked
	// complete.
	IsCompleted bool
	// SkippedPrereqs are the indices and titles of
	// prerequisite steps that have not been completed.
	SkippedPrereqs []PrereqWarning
}

// PrereqWarning describes a skipped prerequisite.
type PrereqWarning struct {
	// StepIndex is the index of the skipped prerequisite.
	StepIndex int
	// Title is the title of the skipped tutorial.
	Title string
}

// GeneratePath builds a tailored learning path from an
// activity profile and a set of tutorials. The path is
// ordered by relevance: tutorials matching strong-confidence
// layers come first, followed by inferred layers.
// Annotations are tailored to the user's stated activities.
func GeneratePath(
	profile *roles.ActivityProfile,
	tutorials []Tutorial,
	schemaVersion string,
) *LearningPath {
	if profile == nil || len(tutorials) == 0 {
		return &LearningPath{
			CompletedSteps: make(map[int]bool),
		}
	}

	roleName := ""
	if profile.Role != nil {
		roleName = profile.Role.Name
	}

	// Build a priority map: layer -> priority index.
	// Lower index = higher priority.
	layerPriority := make(map[int]int)
	for i, lm := range profile.ResolvedLayers {
		layerPriority[lm.Layer] = i
	}

	// Filter tutorials to those matching resolved layers.
	type scoredTutorial struct {
		tutorial Tutorial
		priority int
		mismatch *VersionMismatch
	}

	var scored []scoredTutorial
	resolvedLayers := profile.UniqueLayerNumbers()

	for _, tut := range tutorials {
		if !layerInList(tut.Layer, resolvedLayers) {
			continue
		}
		pri, ok := layerPriority[tut.Layer]
		if !ok {
			pri = len(profile.ResolvedLayers)
		}
		var mm *VersionMismatch
		if schemaVersion != "" &&
			tut.SchemaVersion != "" &&
			tut.SchemaVersion != schemaVersion {
			mm = &VersionMismatch{
				Tutorial:        tut,
				TutorialVersion: tut.SchemaVersion,
				SelectedVersion: schemaVersion,
			}
		}
		scored = append(scored, scoredTutorial{
			tutorial: tut,
			priority: pri,
			mismatch: mm,
		})
	}

	// Sort by priority (strong-confidence layers first),
	// then by layer number.
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].priority != scored[j].priority {
			return scored[i].priority < scored[j].priority
		}
		return scored[i].tutorial.Layer <
			scored[j].tutorial.Layer
	})

	// Build path steps with tailored annotations.
	allKeywords := profile.AllKeywords()
	steps := make([]PathStep, len(scored))
	for i, s := range scored {
		// Compute section-level relevance using the
		// user's activity keywords.
		layerKW := keywordsForLayer(
			s.tutorial.Layer, profile,
		)
		sectionScores := ScoreSections(
			s.tutorial.Sections, layerKW,
		)
		primarySects := PrimarySections(
			s.tutorial.Sections, allKeywords,
		)

		steps[i] = PathStep{
			Tutorial: s.tutorial,
			Layer:    s.tutorial.Layer,
			WhyAnnotation: generateWhyAnnotation(
				s.tutorial, profile,
			),
			HowAnnotation: generateHowAnnotation(
				s.tutorial, profile,
			),
			WhatAnnotation: generateWhatAnnotation(
				s.tutorial, primarySects,
			),
			VersionMismatch:  s.mismatch,
			SectionRelevance: sectionScores,
			PrimarySections:  primarySects,
		}

		// Add prerequisites: earlier steps in the same
		// or lower layer are prerequisites.
		for j := 0; j < i; j++ {
			if scored[j].tutorial.Layer <=
				s.tutorial.Layer {
				steps[i].Prerequisites = append(
					steps[i].Prerequisites, j,
				)
			}
		}
	}

	return &LearningPath{
		TargetRole:     roleName,
		Steps:          steps,
		CompletedSteps: make(map[int]bool),
	}
}

// StepStatus returns navigation info for a step, including
// warnings about skipped prerequisites.
func StepStatus(
	path *LearningPath,
	stepIdx int,
) *StepNavInfo {
	if stepIdx < 0 || stepIdx >= len(path.Steps) {
		return nil
	}

	step := path.Steps[stepIdx]
	info := &StepNavInfo{
		StepIndex:   stepIdx,
		IsCompleted: path.CompletedSteps[stepIdx],
	}

	for _, prereqIdx := range step.Prerequisites {
		if !path.CompletedSteps[prereqIdx] &&
			prereqIdx < len(path.Steps) {
			info.SkippedPrereqs = append(
				info.SkippedPrereqs,
				PrereqWarning{
					StepIndex: prereqIdx,
					Title: path.Steps[prereqIdx].
						Tutorial.Title,
				},
			)
		}
	}

	return info
}

// MissingLayerMessage returns an informative message for
// Gemara layers that have no tutorials available.
func MissingLayerMessage(layer int) string {
	layerNames := map[int]string{
		1: "Guidance",
		2: "Threats & Controls",
		3: "Risk & Policy",
		4: "Sensitive Activities",
		5: "Evaluation",
		6: "Data Collection",
		7: "Reporting",
	}
	name := layerNames[layer]
	if name == "" {
		name = fmt.Sprintf("Layer %d", layer)
	}

	return fmt.Sprintf(
		"No tutorials are currently available for "+
			"Layer %d (%s). Refer to the Gemara model "+
			"documentation for this layer. The closest "+
			"available tutorials have been included in "+
			"your learning path.",
		layer, name,
	)
}

// generateWhyAnnotation creates a "Why this matters" text
// tailored to the user's activities.
func generateWhyAnnotation(
	tut Tutorial,
	profile *roles.ActivityProfile,
) string {
	// Find keywords from the profile that relate to this
	// tutorial's layer.
	relevantKeywords := keywordsForLayer(
		tut.Layer, profile,
	)

	if len(relevantKeywords) > 0 {
		return fmt.Sprintf(
			"Based on your activities (%s), "+
				"understanding %s is essential for "+
				"your work at Gemara Layer %d.",
			strings.Join(relevantKeywords, ", "),
			tut.Title,
			tut.Layer,
		)
	}

	roleName := "your role"
	if profile.Role != nil {
		roleName = profile.Role.Name
	}
	return fmt.Sprintf(
		"As a %s, %s provides foundational "+
			"knowledge for Gemara Layer %d.",
		roleName, tut.Title, tut.Layer,
	)
}

// generateHowAnnotation creates a "How you will use this"
// text tailored to the user's activities.
func generateHowAnnotation(
	tut Tutorial,
	profile *roles.ActivityProfile,
) string {
	relevantKeywords := keywordsForLayer(
		tut.Layer, profile,
	)

	if len(relevantKeywords) > 0 {
		return fmt.Sprintf(
			"Apply the patterns from %s directly to "+
				"your %s workflows.",
			tut.Title,
			strings.Join(relevantKeywords, " and "),
		)
	}

	return fmt.Sprintf(
		"Use the concepts from %s to inform your "+
			"approach to Gemara artifacts at Layer %d.",
		tut.Title, tut.Layer,
	)
}

// generateWhatAnnotation creates a "What you will learn"
// summary, highlighting sections relevant to the user's
// activities when keywords are available.
func generateWhatAnnotation(
	tut Tutorial,
	primarySections []string,
) string {
	if len(tut.Sections) == 0 {
		return fmt.Sprintf(
			"Learn the core concepts of %s.",
			tut.Title,
		)
	}

	if len(primarySections) > 0 {
		// Build the "also covers" list by excluding
		// primary sections.
		primarySet := make(map[string]bool)
		for _, s := range primarySections {
			primarySet[s] = true
		}
		var others []string
		for _, s := range tut.Sections {
			if !primarySet[s] {
				others = append(others, s)
			}
		}
		result := fmt.Sprintf(
			"Focus on: %s.",
			strings.Join(primarySections, ", "),
		)
		if len(others) > 0 {
			result += fmt.Sprintf(
				" Also covers: %s.",
				strings.Join(others, ", "),
			)
		}
		return result
	}

	return fmt.Sprintf(
		"Covers: %s.",
		strings.Join(tut.Sections, ", "),
	)
}

// ScoreSections assigns relevance scores to tutorial
// sections based on keyword matching against section
// headings. Each keyword that appears as a substring in
// a section heading (case-insensitive) adds 2 to the
// section's score.
func ScoreSections(
	sections []string,
	keywords []string,
) map[string]int {
	scores := make(map[string]int)
	for _, section := range sections {
		lower := strings.ToLower(section)
		for _, kw := range keywords {
			if strings.Contains(
				lower,
				strings.ToLower(kw),
			) {
				scores[section] += 2
			}
		}
	}
	return scores
}

// PrimarySections returns sections with non-zero relevance
// scores, ordered by score descending.
func PrimarySections(
	sections []string,
	keywords []string,
) []string {
	scores := ScoreSections(sections, keywords)

	type scored struct {
		name  string
		score int
	}
	var matched []scored
	for _, s := range sections {
		if scores[s] > 0 {
			matched = append(matched, scored{
				name:  s,
				score: scores[s],
			})
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].score > matched[j].score
	})

	result := make([]string, len(matched))
	for i, m := range matched {
		result[i] = m.name
	}
	return result
}

// keywordsForLayer returns the profile keywords that map to
// the given layer.
func keywordsForLayer(
	layer int,
	profile *roles.ActivityProfile,
) []string {
	var relevant []string
	for _, lm := range profile.ResolvedLayers {
		if lm.Layer == layer {
			relevant = append(
				relevant, lm.Keywords...,
			)
		}
	}
	return relevant
}

// layerInList checks if a layer is in the list.
func layerInList(layer int, layers []int) bool {
	for _, l := range layers {
		if l == layer {
			return true
		}
	}
	return false
}
