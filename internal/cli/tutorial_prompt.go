// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/hbraswelrh/pacman/internal/blocks"
	"github.com/hbraswelrh/pacman/internal/tutorials"
)

// TutorialPromptConfig holds dependencies for the
// interactive tutorial player.
type TutorialPromptConfig struct {
	// Prompter handles user interaction.
	Prompter FreeTextPrompter
	// LearningPath is the tailored learning path.
	LearningPath *tutorials.LearningPath
	// TutorialsDir is the path to tutorial files.
	TutorialsDir string
	// RoleName is the user's role for personalization.
	RoleName string
	// Keywords are the user's activity keywords for
	// content block retrieval.
	Keywords []string
}

// TutorialPromptResult holds the outcome of a tutorial
// session.
type TutorialPromptResult struct {
	// CompletedSteps tracks which path steps were marked
	// complete during this session.
	CompletedSteps map[int]bool
}

// RunTutorialPlayer presents the learning path and lets
// the user select and walk through tutorials section by
// section.
func RunTutorialPlayer(
	cfg *TutorialPromptConfig,
	out io.Writer,
) (*TutorialPromptResult, error) {
	path := cfg.LearningPath
	if path == nil || len(path.Steps) == 0 {
		fmt.Fprintln(out, RenderNote(
			"No tutorials in your learning path. "+
				"Try different activity keywords.",
		))
		return &TutorialPromptResult{
			CompletedSteps: make(map[int]bool),
		}, nil
	}

	for {
		// Show step selection menu.
		stepIdx, err := selectStep(cfg, path, out)
		if err != nil {
			return nil, fmt.Errorf(
				"select step: %w", err,
			)
		}
		if stepIdx < 0 {
			// User chose to return to main menu.
			break
		}

		// Run the selected tutorial.
		completed, err := runTutorialStep(
			cfg, path, stepIdx, out,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"tutorial step: %w", err,
			)
		}
		if completed {
			path.CompletedSteps[stepIdx] = true
		}
	}

	return &TutorialPromptResult{
		CompletedSteps: path.CompletedSteps,
	}, nil
}

// selectStep presents the learning path steps and returns
// the selected index, or -1 to exit.
func selectStep(
	cfg *TutorialPromptConfig,
	path *tutorials.LearningPath,
	out io.Writer,
) (int, error) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, headingStyle.Render(
		"Your Tutorials",
	))
	if cfg.RoleName != "" {
		fmt.Fprintln(out, " "+orangeStyle.Render(
			"("+cfg.RoleName+")",
		))
	}
	fmt.Fprintln(out)

	options := make([]string, len(path.Steps)+1)
	for i, step := range path.Steps {
		status := "  "
		if path.CompletedSteps[i] {
			status = successStyle.Render("✓ ")
		}
		layerLabel := renderLayerName(step.Layer)
		options[i] = fmt.Sprintf(
			"%s%s — %s",
			status, step.Tutorial.Title, layerLabel,
		)
	}
	options[len(path.Steps)] = "Back to main menu"

	choice, err := cfg.Prompter.Ask(
		"Select a tutorial to start:",
		options,
	)
	if err != nil {
		return -1, err
	}
	if choice >= len(path.Steps) {
		return -1, nil
	}
	return choice, nil
}

// runTutorialStep runs a single tutorial step, walking
// through its sections one at a time. Returns true if the
// user marked it complete.
func runTutorialStep(
	cfg *TutorialPromptConfig,
	path *tutorials.LearningPath,
	stepIdx int,
	out io.Writer,
) (bool, error) {
	step := path.Steps[stepIdx]

	// Parse sections from the tutorial file.
	sections, err := tutorials.ParseSections(
		step.Tutorial.FilePath,
	)
	if err != nil {
		fmt.Fprintln(out, RenderWarning(
			"Could not parse tutorial sections: "+
				err.Error(),
		))
		return false, nil
	}

	if len(sections) == 0 {
		fmt.Fprintln(out, RenderNote(
			"This tutorial has no sections yet.",
		))
		return false, nil
	}

	// Display tutorial header.
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderTutorialHeader(
		step.Tutorial.Title,
		step.Layer,
		step.WhyAnnotation,
		len(sections),
	))

	// Surface related wizards for Layer 2 tutorials.
	wizards := AvailableWizards()
	var relatedWizards []WizardInfo
	for _, w := range wizards {
		if w.Layer == step.Layer {
			relatedWizards = append(relatedWizards, w)
		}
	}
	if len(relatedWizards) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, faintStyle.Render(
			"Related MCP wizards available:"))
		for _, w := range relatedWizards {
			fmt.Fprintf(out, "  %s %s\n",
				successStyle.Render("→"),
				faintStyle.Render(w.Title+
					" — "+w.Description),
			)
		}
		fmt.Fprintln(out, faintStyle.Render(
			"  Select 'Launch a wizard' from the "+
				"main menu to start one."))
	}

	// Extract content blocks from the tutorial for
	// inline surfacing of related patterns.
	tutBlocks := blocks.ExtractBlocks(
		step.Tutorial,
		sections,
		step.Tutorial.SchemaVersion,
	)
	blockIndex := blocks.NewBlockIndex(tutBlocks)

	// Walk through sections.
	sectionIdx := 0
	for {
		section := sections[sectionIdx]
		totalSections := len(sections)

		fmt.Fprintln(out)
		fmt.Fprintln(out, RenderTutorialSection(
			section.Heading,
			section.Body,
			sectionIdx,
			totalSections,
			step.Tutorial.Title,
			cfg.RoleName,
		))

		// Surface related content blocks for the
		// current section based on the user's
		// activity keywords.
		if len(cfg.Keywords) > 0 && blockIndex != nil {
			related := blocks.RetrieveBlocks(
				blockIndex,
				[]int{step.Layer},
				cfg.Keywords,
			)
			if len(related) > 0 {
				fmt.Fprintln(out)
				fmt.Fprintln(out, faintStyle.Render(
					"  Related patterns:"))
				limit := len(related)
				if limit > 2 {
					limit = 2
				}
				for i := range related[:limit] {
					adapt := blocks.GenerateAdaptation(
						&related[i].Block,
						strings.Join(
							cfg.Keywords, ", ",
						),
					)
					fmt.Fprintln(out,
						"  "+faintStyle.Render(
							"  "+adapt,
						),
					)
				}
			}
		}
		fmt.Fprintln(out)

		// Build navigation options.
		navOpts := buildNavOptions(
			sectionIdx, totalSections,
		)

		navChoice, err := cfg.Prompter.Ask(
			"", navOpts,
		)
		if err != nil {
			return false, err
		}

		action := resolveNavAction(
			navChoice, sectionIdx, totalSections,
		)

		switch action {
		case navNext:
			sectionIdx++
		case navPrev:
			sectionIdx--
		case navComplete:
			fmt.Fprintln(out, RenderSuccess(
				fmt.Sprintf(
					"Completed: %s",
					step.Tutorial.Title,
				),
			))
			return true, nil
		case navBack:
			return false, nil
		}
	}
}

// Navigation actions.
const (
	navNext     = "next"
	navPrev     = "prev"
	navComplete = "complete"
	navBack     = "back"
)

// buildNavOptions constructs the navigation menu for the
// current section position.
func buildNavOptions(
	current int,
	total int,
) []string {
	var opts []string
	if current < total-1 {
		opts = append(opts, "Next section")
	}
	if current > 0 {
		opts = append(opts, "Previous section")
	}
	if current == total-1 {
		opts = append(opts, "Mark complete")
	}
	opts = append(opts, "Back to tutorial list")
	return opts
}

// resolveNavAction maps a choice index to a navigation
// action based on the current section position.
func resolveNavAction(
	choice int,
	current int,
	total int,
) string {
	opts := buildNavOptions(current, total)
	if choice < 0 || choice >= len(opts) {
		return navBack
	}
	label := opts[choice]
	switch {
	case strings.HasPrefix(label, "Next"):
		return navNext
	case strings.HasPrefix(label, "Previous"):
		return navPrev
	case strings.HasPrefix(label, "Mark"):
		return navComplete
	default:
		return navBack
	}
}

// renderLayerName returns a short layer label.
func renderLayerName(layer int) string {
	names := map[int]string{
		1: "Guidance",
		2: "Threats & Controls",
		3: "Risk & Policy",
		4: "Sensitive Activities",
		5: "Evaluation",
		6: "Data Collection",
		7: "Reporting",
	}
	if name, ok := names[layer]; ok {
		return fmt.Sprintf("L%d %s", layer, name)
	}
	return fmt.Sprintf("L%d", layer)
}
