// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/hbraswelrh/gemara-user-journey/internal/blocks"
	"github.com/hbraswelrh/gemara-user-journey/internal/schema"
	"github.com/hbraswelrh/gemara-user-journey/internal/session"
	"github.com/hbraswelrh/gemara-user-journey/internal/tutorials"
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
	// Session is the current session state, used for
	// building handoff summaries after tutorial
	// completion.
	Session *session.Session
	// SelectionResult holds version selection metadata
	// for handoff summary rendering (experimental
	// schemas, compatibility warnings).
	SelectionResult *schema.SelectionResult
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
// section with focused questions at each step.
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
		stepIdx, err := selectStep(cfg, path, out)
		if err != nil {
			return nil, fmt.Errorf(
				"select step: %w", err,
			)
		}
		if stepIdx < 0 {
			break
		}

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
		fmt.Fprintln(out, " "+roleInfoStyle.Render(
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

// runTutorialStep runs a single tutorial with question-
// focused section-by-section walkthrough.
func runTutorialStep(
	cfg *TutorialPromptConfig,
	path *tutorials.LearningPath,
	stepIdx int,
	out io.Writer,
) (bool, error) {
	step := path.Steps[stepIdx]

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

	// Surface related wizards.
	wizards := AvailableWizards()
	for _, w := range wizards {
		if w.Layer == step.Layer {
			fmt.Fprintln(out)
			fmt.Fprintf(out, "  %s %s\n",
				successStyle.Render("→"),
				faintStyle.Render(
					"MCP Wizard: "+w.Title+
						" — use 'Launch a wizard' "+
						"from the main menu",
				),
			)
		}
	}

	// Show focus sections if the user's activities match
	// specific sections within this tutorial.
	if len(step.PrimarySections) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, subtleStyle.Render(
			"Based on your activities, your focus "+
				"sections are:",
		))
		for _, s := range step.PrimarySections {
			fmt.Fprintf(out, "  %s %s\n",
				successStyle.Render("★"),
				s,
			)
		}
		otherCount := len(sections) -
			len(step.PrimarySections)
		if otherCount > 0 {
			fmt.Fprintln(out, faintStyle.Render(
				fmt.Sprintf(
					"  + %d additional sections",
					otherCount,
				),
			))
		}
	}

	// Extract content blocks for inline hints.
	tutBlocks := blocks.ExtractBlocks(
		step.Tutorial,
		sections,
		step.Tutorial.SchemaVersion,
	)
	blockIndex := blocks.NewBlockIndex(tutBlocks)

	// Walk through sections with questions.
	sectionIdx := 0
	for {
		section := sections[sectionIdx]
		total := len(sections)

		// Phase 1: Show section intro and ask a
		// focused question.
		intro, detail := SplitSectionBody(section.Body)

		fmt.Fprintln(out)
		fmt.Fprintln(out, renderSectionIntro(
			section.Heading, intro,
			sectionIdx, total,
		))

		// Show relevance badge for sections that match
		// the user's stated activities.
		if score, ok := step.SectionRelevance[section.Heading]; ok && score > 0 {
			matchedKW := findMatchingKeywords(
				section.Heading, cfg.Keywords,
			)
			if len(matchedKW) > 0 {
				fmt.Fprintln(out, successStyle.Render(
					"  ★ Matches your activity: "+
						strings.Join(matchedKW, ", "),
				))
			}
		}

		// Generate and ask a focused question.
		question := generateSectionQuestion(
			section.Heading,
			step.Tutorial.Title,
			cfg.RoleName,
		)
		qOpts := generateQuestionOptions(
			section.Heading,
			step.Tutorial.Title,
		)

		fmt.Fprintln(out)
		qChoice, err := cfg.Prompter.Ask(
			question, qOpts,
		)
		if err != nil {
			return false, err
		}

		// Run follow-up prompts based on the user's
		// answer to deepen understanding.
		err = runFollowUps(
			cfg, section.Heading, qChoice,
			qOpts, out,
		)
		if err != nil {
			return false, err
		}

		// Phase 2: Show the full detail content.
		if detail != "" {
			fmt.Fprintln(out)
			fmt.Fprintln(out, renderSectionDetail(
				detail, cfg.RoleName,
			))
		}

		// Phase 3: Surface related content blocks.
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

		// Phase 4: Navigation with application prompt.
		fmt.Fprintln(out)
		navOpts := buildNavOptions(sectionIdx, total)

		navChoice, err := cfg.Prompter.Ask(
			"", navOpts,
		)
		if err != nil {
			return false, err
		}

		action := resolveNavAction(
			navChoice, sectionIdx, total,
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

			// Render handoff summary directing the
			// user to OpenCode with gemara-mcp for
			// authoring.
			if cfg.Session != nil {
				summary := BuildHandoffSummary(
					&step,
					cfg.Session,
					cfg.SelectionResult,
				)
				RenderHandoffSummary(summary, out)
			}

			return true, nil
		case navBack:
			return false, nil
		}
	}
}

// runFollowUps asks targeted follow-up questions based on
// the user's answer to the main section question. This
// ensures all paths are explored — e.g., choosing "Import
// from FINOS CCC" still prompts about custom items.
func runFollowUps(
	cfg *TutorialPromptConfig,
	heading string,
	choice int,
	options []string,
	out io.Writer,
) error {
	lower := strings.ToLower(heading)
	if choice < 0 || choice >= len(options) {
		return nil
	}
	selected := strings.ToLower(options[choice])

	// --- Capability Identification follow-ups ---
	if strings.Contains(lower, "capability") &&
		strings.Contains(lower, "ident") {
		return followUpCapabilities(
			cfg, selected, out,
		)
	}

	// --- Threat Identification follow-ups ---
	if strings.Contains(lower, "threat") &&
		strings.Contains(lower, "ident") {
		return followUpThreats(cfg, selected, out)
	}

	// --- Control Structure / Authoring follow-ups ---
	if strings.Contains(lower, "control") &&
		(strings.Contains(lower, "structure") ||
			strings.Contains(lower, "authoring") ||
			strings.Contains(lower, "custom")) {
		return followUpControls(cfg, selected, out)
	}

	// --- Importing External Catalogs follow-ups ---
	if strings.Contains(lower, "importing") {
		return followUpImporting(cfg, selected, out)
	}

	return nil
}

func followUpCapabilities(
	cfg *TutorialPromptConfig,
	selected string,
	out io.Writer,
) error {
	roleCtx := ""
	if cfg.RoleName != "" {
		roleCtx = " for your " + cfg.RoleName + " work"
	}

	if strings.Contains(selected, "import") &&
		!strings.Contains(selected, "custom") &&
		!strings.Contains(selected, "both") {
		// User chose import only — ask about custom.
		fmt.Fprintln(out)
		fmt.Fprintln(out, stepBarStyle.Render(
			subtleStyle.Render(
				"FINOS CCC Core provides common "+
					"capabilities like Active "+
					"Ingestion (CP29) and Resource "+
					"Versioning (CP18). But your "+
					"component likely has unique "+
					"capabilities too.",
			),
		))

		_, err := cfg.Prompter.Ask(
			"Would you also like to define custom "+
				"capabilities specific to your "+
				"component"+roleCtx+"?",
			[]string{
				"Yes, define custom capabilities too",
				"No, imported capabilities are sufficient",
				"Show me what custom capabilities look like",
			},
		)
		return err
	}

	if strings.Contains(selected, "custom") &&
		!strings.Contains(selected, "import") &&
		!strings.Contains(selected, "both") {
		// User chose custom only — suggest imports.
		fmt.Fprintln(out)
		fmt.Fprintln(out, stepBarStyle.Render(
			subtleStyle.Render(
				"Before defining everything from "+
					"scratch, consider importing from "+
					"FINOS CCC Core. It provides "+
					"well-vetted capabilities that "+
					"apply broadly across cloud "+
					"services and can save time.",
			),
		))

		_, err := cfg.Prompter.Ask(
			"Would you also like to import "+
				"capabilities from FINOS CCC Core?",
			[]string{
				"Yes, import from CCC and add custom",
				"No, custom capabilities only",
			},
		)
		return err
	}

	return nil
}

func followUpThreats(
	cfg *TutorialPromptConfig,
	selected string,
	out io.Writer,
) error {
	roleCtx := ""
	if cfg.RoleName != "" {
		roleCtx = " in your " + cfg.RoleName + " role"
	}

	if strings.Contains(selected, "import") &&
		!strings.Contains(selected, "custom") &&
		!strings.Contains(selected, "both") {
		// User chose imported threats only.
		fmt.Fprintln(out)
		fmt.Fprintln(out, stepBarStyle.Render(
			subtleStyle.Render(
				"CCC Core threats cover common "+
					"attack patterns, but your "+
					"component will have unique "+
					"threat scenarios that imported "+
					"catalogs don't address.",
			),
		))

		_, err := cfg.Prompter.Ask(
			"Would you also like to define custom "+
				"threats for your component"+
				roleCtx+"?",
			[]string{
				"Yes, define custom threats too",
				"No, imported threats cover my scope",
				"Show me the custom threat structure",
			},
		)
		if err != nil {
			return err
		}

		// MITRE ATT&CK follow-up.
		fmt.Fprintln(out)
		_, err = cfg.Prompter.Ask(
			"Would you like to link threats to "+
				"MITRE ATT&CK techniques? This adds "+
				"structured vector entries for threat "+
				"intelligence.",
			[]string{
				"Yes, link to MITRE ATT&CK",
				"No, skip MITRE ATT&CK linking",
				"What is MITRE ATT&CK linking?",
			},
		)
		return err
	}

	if strings.Contains(selected, "custom") &&
		!strings.Contains(selected, "import") &&
		!strings.Contains(selected, "both") {
		// User chose custom only — suggest imports.
		fmt.Fprintln(out)
		fmt.Fprintln(out, stepBarStyle.Render(
			subtleStyle.Render(
				"Before defining all threats from "+
					"scratch, check if CCC Core "+
					"already defines threats for "+
					"the capabilities you imported. "+
					"For example, TH14 (Older "+
					"Resource Versions) maps to "+
					"CP18 (Resource Versioning).",
			),
		))

		_, err := cfg.Prompter.Ask(
			"Would you also like to check for "+
				"imported threats from CCC Core?",
			[]string{
				"Yes, import matching CCC threats",
				"No, define all threats custom",
			},
		)
		return err
	}

	// For "both" or "show structure" — ask about MITRE.
	fmt.Fprintln(out)
	_, err := cfg.Prompter.Ask(
		"Would you like to link threats to "+
			"MITRE ATT&CK techniques?",
		[]string{
			"Yes, link to MITRE ATT&CK",
			"No, skip MITRE ATT&CK linking",
			"What is MITRE ATT&CK linking?",
		},
	)
	return err
}

func followUpControls(
	cfg *TutorialPromptConfig,
	selected string,
	out io.Writer,
) error {
	// After any control-related answer, prompt about
	// assessment requirements.
	fmt.Fprintln(out)
	fmt.Fprintln(out, stepBarStyle.Render(
		subtleStyle.Render(
			"Each control needs assessment "+
				"requirements — testable conditions "+
				"that evaluators use to determine "+
				"pass or fail. Use the pattern:\n\n"+
				"  \"When [condition], [subject] MUST "+
				"[observable action].\"\n\n"+
				"Good: \"When YAML is submitted, the "+
				"server MUST reject payloads exceeding "+
				"a configured maximum size.\"\n"+
				"Bad: \"User input MUST be validated.\"",
		),
	))

	_, err := cfg.Prompter.Ask(
		"How would you like to define assessment "+
			"requirements?",
		[]string{
			"Write requirements per control",
			"Start with examples and adapt",
			"Show me the requirement format",
		},
	)
	if err != nil {
		return err
	}

	// Follow-up: threat mapping.
	fmt.Fprintln(out)
	_, err = cfg.Prompter.Ask(
		"Would you like to map controls to "+
			"threats from your Threat Catalog?",
		[]string{
			"Yes, link controls to threat IDs",
			"No, define controls without threat mapping",
			"Show me how threat mapping works",
		},
	)
	return err
}

func followUpImporting(
	cfg *TutorialPromptConfig,
	selected string,
	out io.Writer,
) error {
	if strings.Contains(selected, "finos") ||
		strings.Contains(selected, "ccc") {
		fmt.Fprintln(out)
		fmt.Fprintln(out, stepBarStyle.Render(
			subtleStyle.Render(
				"FINOS CCC Core provides pre-built "+
					"controls with threat mappings. "+
					"You can import these and add "+
					"your own custom controls on top.",
			),
		))

		_, err := cfg.Prompter.Ask(
			"Would you also like to define custom "+
				"controls specific to your component?",
			[]string{
				"Yes, import CCC and add custom controls",
				"No, imported controls are sufficient",
				"Show me how to combine imports and custom",
			},
		)
		return err
	}

	if strings.Contains(selected, "osps") {
		fmt.Fprintln(out)
		_, err := cfg.Prompter.Ask(
			"OSPS Baseline focuses on open source "+
				"project security. Would you also "+
				"like to import from FINOS CCC Core "+
				"for cloud-specific controls?",
			[]string{
				"Yes, import from both OSPS and CCC",
				"No, OSPS Baseline only",
			},
		)
		return err
	}

	return nil
}

// SplitSectionBody splits a section body into an intro
// (first paragraph) and detail (everything after). This
// enables progressive disclosure — show the intro first,
// then reveal details after the user engages.
func SplitSectionBody(body string) (string, string) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "", ""
	}

	// Split on double newline (paragraph break).
	parts := strings.SplitN(body, "\n\n", 2)
	intro := strings.TrimSpace(parts[0])
	detail := ""
	if len(parts) > 1 {
		detail = strings.TrimSpace(parts[1])
	}
	return intro, detail
}

// renderSectionIntro renders the section heading and
// introductory paragraph with progress.
func renderSectionIntro(
	heading string,
	intro string,
	current int,
	total int,
) string {
	progress := fmt.Sprintf(
		"Section %d of %d", current+1, total,
	)

	var lines []string
	lines = append(lines,
		stepNumStyle.Render(progress)+"  "+
			headingStyle.Render(heading),
	)
	lines = append(lines, "")
	if intro != "" {
		rendered := RenderMarkdown(intro)
		lines = append(lines, rendered)
	}

	return stepBarStyle.Render(
		strings.Join(lines, "\n"),
	)
}

// renderSectionDetail renders the detailed content after
// the user has engaged with the introductory question.
func renderSectionDetail(
	detail string,
	roleName string,
) string {
	var lines []string

	rendered := RenderMarkdown(detail)
	lines = append(lines, rendered)

	if roleName != "" {
		lines = append(lines, "")
		lines = append(lines,
			annotationLabelStyle.Render("Apply: ")+
				annotationTextStyle.Render(
					"Consider how this applies to "+
						"your work as a "+roleName+".",
				),
		)
	}

	return stepBarStyle.Render(
		strings.Join(lines, "\n"),
	)
}

// generateSectionQuestion creates a focused question for
// the section topic to help the user engage with the
// material before seeing the detailed content.
func generateSectionQuestion(
	heading string,
	tutorialTitle string,
	roleName string,
) string {
	lower := strings.ToLower(heading)

	roleCtx := ""
	if roleName != "" {
		roleCtx = " for your role as a " + roleName
	}

	switch {
	// Threat Assessment sections
	case strings.Contains(lower, "scope"):
		return "What component or technology are you " +
			"looking to assess" + roleCtx + "?"
	case strings.Contains(lower, "capability") &&
		strings.Contains(lower, "ident"):
		return "What are the core functions of the " +
			"component you are assessing" + roleCtx +
			"?"
	case strings.Contains(lower, "threat") &&
		strings.Contains(lower, "ident"):
		return "For each capability, what could go " +
			"wrong" + roleCtx + "?"
	case strings.Contains(lower, "validation") ||
		strings.Contains(lower, "validate"):
		return "How will you validate your artifact " +
			"against the Gemara schema?"

	// Guidance Catalog sections
	case strings.Contains(lower, "creating") &&
		strings.Contains(lower, "guidance"):
		return "What type of guidance catalog best " +
			"fits your organization's needs" +
			roleCtx + "?"
	case strings.Contains(lower, "metadata"):
		return "What metadata fields are essential " +
			"for your artifact?"
	case strings.Contains(lower, "families") ||
		strings.Contains(lower, "groups"):
		return "How would you group your guidelines " +
			"by theme" + roleCtx + "?"
	case strings.Contains(lower, "cross-ref"):
		return "Which guidelines in your catalog " +
			"relate to each other?"
	case strings.Contains(lower, "mapping"):
		return "Which external standards do you need " +
			"to map your guidance to?"

	// Control Catalog sections
	case strings.Contains(lower, "control") &&
		strings.Contains(lower, "structure"):
		return "What applicability categories apply " +
			"to your controls (e.g., production, " +
			"CI/CD)?"
	case strings.Contains(lower, "custom control") ||
		strings.Contains(lower, "authoring"):
		return "What risks need to be reduced by " +
			"your controls" + roleCtx + "?"
	case strings.Contains(lower, "importing"):
		return "Which external control frameworks " +
			"would you like to import from?"
	case strings.Contains(lower, "osps"):
		return "Does your project align with OpenSSF " +
			"security baseline requirements?"
	case strings.Contains(lower, "finos") ||
		strings.Contains(lower, "ccc"):
		return "Would you like to import controls " +
			"from the FINOS CCC Core catalog?"

	// Policy sections
	case strings.Contains(lower, "policy") &&
		strings.Contains(lower, "structure"):
		return "What is the scope and purpose of " +
			"your policy" + roleCtx + "?"
	case strings.Contains(lower, "implementation"):
		return "What is your timeline for rolling " +
			"out this policy?"
	case strings.Contains(lower, "evaluation") &&
		strings.Contains(lower, "timeline"):
		return "How will you evaluate compliance " +
			"during the assessment phase?"
	case strings.Contains(lower, "enforcement"):
		return "When should enforcement begin and " +
			"what are the consequences?"
	case strings.Contains(lower, "adherence"):
		return "How will adherence be measured and " +
			"what happens on non-compliance?"

	default:
		return "What do you want to learn about " +
			heading + roleCtx + "?"
	}
}

// generateQuestionOptions creates multiple-choice options
// that help the user think about the section topic before
// diving into the details.
func generateQuestionOptions(
	heading string,
	tutorialTitle string,
) []string {
	lower := strings.ToLower(heading)

	switch {
	case strings.Contains(lower, "scope"):
		return []string{
			"A specific service or API",
			"An infrastructure component",
			"A CI/CD pipeline or build system",
			"A cloud platform feature",
			"I'm not sure yet — show me examples",
		}
	case strings.Contains(lower, "capability") &&
		strings.Contains(lower, "ident"):
		return []string{
			"Import from FINOS CCC Core catalog",
			"Define custom capabilities",
			"Both import and custom",
			"Show me examples first",
		}
	case strings.Contains(lower, "threat") &&
		strings.Contains(lower, "ident"):
		return []string{
			"Check for imported threats (CCC)",
			"Define custom threats for my component",
			"Both imported and custom threats",
			"Show me the threat structure first",
		}
	case strings.Contains(lower, "validation") ||
		strings.Contains(lower, "validate"):
		return []string{
			"Validate with cue vet (local)",
			"Validate via MCP server",
			"Show me the validation commands",
		}

	case strings.Contains(lower, "creating") &&
		strings.Contains(lower, "guidance"):
		return []string{
			"Standard (ISO, PCI, NIST)",
			"Regulation (HIPAA, GDPR, CRA)",
			"Best Practice (internal, OWASP)",
			"Framework (NIST CSF)",
			"Show me the options",
		}
	case strings.Contains(lower, "metadata"):
		return []string{
			"Set up with mapping references",
			"Minimal metadata only",
			"Show me the required fields",
		}
	case strings.Contains(lower, "families") ||
		strings.Contains(lower, "groups"):
		return []string{
			"Group by security domain",
			"Group by technology area",
			"Show me examples",
		}

	case strings.Contains(lower, "control") &&
		strings.Contains(lower, "structure"):
		return []string{
			"Production environments only",
			"All deployment environments",
			"CI/CD pipelines",
			"Show me the structure",
		}
	case strings.Contains(lower, "custom control") ||
		strings.Contains(lower, "authoring"):
		return []string{
			"Write controls with assessment requirements",
			"Map controls to existing threats",
			"Show me an example control",
		}
	case strings.Contains(lower, "importing"):
		return []string{
			"FINOS CCC Core",
			"OSPS Baseline",
			"Both CCC and OSPS",
			"Show me how imports work",
		}

	case strings.Contains(lower, "policy") &&
		strings.Contains(lower, "structure"):
		return []string{
			"Cloud and web application security",
			"Supply chain security",
			"Data protection",
			"Show me the structure",
		}
	case strings.Contains(lower, "implementation") ||
		strings.Contains(lower, "timeline"):
		return []string{
			"Immediate enforcement",
			"Phased rollout with evaluation first",
			"Show me the timeline fields",
		}
	case strings.Contains(lower, "adherence"):
		return []string{
			"Automated checks (CI/CD gates)",
			"Manual review process",
			"Combination of automated and manual",
			"Show me adherence options",
		}

	default:
		return []string{
			"Walk me through this section",
			"Show me examples",
			"Skip to the next section",
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
		opts = append(opts, "Continue to next section")
	}
	if current > 0 {
		opts = append(opts, "Go back to previous section")
	}
	if current == total-1 {
		opts = append(opts, "Mark tutorial as complete")
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
	case strings.HasPrefix(label, "Continue"):
		return navNext
	case strings.HasPrefix(label, "Go back"):
		return navPrev
	case strings.HasPrefix(label, "Mark"):
		return navComplete
	default:
		return navBack
	}
}

// findMatchingKeywords returns keywords that match a
// section heading (case-insensitive substring match).
func findMatchingKeywords(
	heading string,
	keywords []string,
) []string {
	lower := strings.ToLower(heading)
	var matched []string
	for _, kw := range keywords {
		if strings.Contains(
			lower, strings.ToLower(kw),
		) {
			matched = append(matched, kw)
		}
	}
	return matched
}

// renderLayerName returns a short layer label using the
// centralized LayerNames map from styles.go.
func renderLayerName(layer int) string {
	names := LayerNames
	if name, ok := names[layer]; ok {
		return fmt.Sprintf("L%d %s", layer, name)
	}
	return fmt.Sprintf("L%d", layer)
}
