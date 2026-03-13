// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"charm.land/lipgloss/v2/table"

	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/team"
	"github.com/hbraswelrh/pacman/internal/tutorials"
)

// Color palette for the Pac-Man TUI. Uses AdaptiveColor to
// support both light and dark terminal backgrounds,
// including OpenCode IDE environments.
var (
	colorPrimary = compat.AdaptiveColor{
		Light: lipgloss.Color("#6B3FD4"),
		Dark:  lipgloss.Color("#7D56F4"),
	}
	colorSecondary = compat.AdaptiveColor{
		Light: lipgloss.Color("#4A2BA0"),
		Dark:  lipgloss.Color("#5B3CC4"),
	}
	colorSubtle = compat.AdaptiveColor{
		Light: lipgloss.Color("#555555"),
		Dark:  lipgloss.Color("#6C6C6C"),
	}
	colorSuccess = compat.AdaptiveColor{
		Light: lipgloss.Color("#038A5A"),
		Dark:  lipgloss.Color("#04B575"),
	}
	colorWarning = compat.AdaptiveColor{
		Light: lipgloss.Color("#CC9900"),
		Dark:  lipgloss.Color("#FFCC00"),
	}
	colorFaint = compat.AdaptiveColor{
		Light: lipgloss.Color("#666666"),
		Dark:  lipgloss.Color("#999999"),
	}
	colorWhite = compat.AdaptiveColor{
		Light: lipgloss.Color("#1A1A1A"),
		Dark:  lipgloss.Color("#FAFAFA"),
	}
	colorDim = compat.AdaptiveColor{
		Light: lipgloss.Color("#AAAAAA"),
		Dark:  lipgloss.Color("#555555"),
	}
	colorCyan = compat.AdaptiveColor{
		Light: lipgloss.Color("#0088AA"),
		Dark:  lipgloss.Color("#00D4FF"),
	}
	colorOrange = compat.AdaptiveColor{
		Light: lipgloss.Color("#CC6600"),
		Dark:  lipgloss.Color("#FF8C00"),
	}
)

// Reusable text styles.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorPrimary).
			Padding(0, 1)

	headingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	subtleStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSuccess)

	warningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWarning)

	faintStyle = lipgloss.NewStyle().
			Foreground(colorFaint)

	sectionDivider = lipgloss.NewStyle().
			Foreground(colorDim)

	layerBadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorSecondary).
			Padding(0, 1)

	stepNumStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan)

	tutorialTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWhite)

	annotationLabelStyle = lipgloss.NewStyle().
				Foreground(colorPrimary)

	annotationTextStyle = lipgloss.NewStyle().
				Foreground(colorFaint)

	confidenceStrongStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSuccess)

	confidenceInferredStyle = lipgloss.NewStyle().
				Foreground(colorFaint).
				Italic(true)

	keywordTagStyle = lipgloss.NewStyle().
			Foreground(colorCyan)

	questionStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)

	answerStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	roleInfoStyle = lipgloss.NewStyle().
			Foreground(colorOrange).
			Bold(true)

	orangeStyle = lipgloss.NewStyle().
			Foreground(colorOrange).
			Bold(true)

	// stepBarStyle uses a left-side colored border only.
	// No box outline — just a vertical accent bar with
	// padding. This wraps cleanly at any terminal width.
	stepBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderRight(false).
			BorderTop(false).
			BorderBottom(false).
			BorderForeground(colorPrimary).
			PaddingLeft(2).
			MarginLeft(2).
			MarginBottom(1)
)

// LayerNames maps Gemara layer numbers to display names.
// Centralized here for use by all TUI components.
var LayerNames = map[int]string{
	consts.LayerGuidance:          "Guidance",
	consts.LayerThreatsControls:   "Threats & Controls",
	consts.LayerRiskPolicy:        "Risk & Policy",
	consts.LayerSensitiveActivity: "Sensitive Activities",
	consts.LayerEvaluation:        "Evaluation",
	consts.LayerDataCollection:    "Data Collection",
	consts.LayerReporting:         "Reporting",
}

// RenderBanner returns the styled application banner.
func RenderBanner() string {
	banner := titleStyle.Render(
		"  Pac-Man — Gemara Tutorial Engine  ",
	)
	return "\n" + banner + "\n"
}

// RenderDivider returns a styled horizontal divider.
func RenderDivider() string {
	return sectionDivider.Render(
		strings.Repeat("─", 52),
	)
}

// RenderMCPToolsPanel returns a styled panel showing
// the available MCP tools in a table layout.
func RenderMCPToolsPanel() string {
	heading := headingStyle.Render(
		"Gemara MCP Server",
	)
	intro := subtleStyle.Render(
		"The MCP server unlocks these tools:",
	)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(
			lipgloss.NewStyle().
				Foreground(colorSubtle),
		).
		StyleFunc(
			func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return headerStyle
				}
				if col == 0 {
					return lipgloss.NewStyle().
						Bold(true).
						Foreground(colorWhite)
				}
				return faintStyle
			},
		).
		Headers("Tool", "Description").
		Row(
			consts.ToolGetLexicon,
			"Lexicon lookups",
		).
		Row(
			consts.ToolValidateArtifact,
			"Schema validation",
		).
		Row(
			consts.ToolGetSchemaDocs,
			"Schema docs reference",
		)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		heading,
		intro,
		"",
		t.String(),
		"",
	)
}

// RenderSessionStatus returns a styled session summary.
// Uses simple indented text — no box borders.
func RenderSessionStatus(
	version string,
	fallback bool,
) string {
	header := headingStyle.Render("Session Ready")

	vLabel := faintStyle.Render("Schema version:")
	vValue := successStyle.Render(version)
	if version == "" {
		vValue = faintStyle.Render("(not selected)")
	}
	versionLine := "  " + vLabel + " " + vValue

	mLabel := faintStyle.Render("Fallback mode: ")
	modeLabel := successStyle.Render("no")
	if fallback {
		modeLabel = warningStyle.Render("yes")
	}
	fallbackLine := "  " + mLabel + " " + modeLabel

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		RenderDivider(),
		"",
		header,
		"",
		versionLine,
		fallbackLine,
		"",
	)
}

// RenderSessionRoleInfo returns a styled role/path summary
// to display after the session status.
func RenderSessionRoleInfo(
	roleName string,
	pathSteps int,
) string {
	if roleName == "" {
		return ""
	}

	rLabel := faintStyle.Render("  Role:")
	rValue := roleInfoStyle.Render(roleName)

	pLabel := faintStyle.Render("  Learning path:")
	pValue := successStyle.Render(
		fmt.Sprintf("%d steps", pathSteps),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		rLabel+" "+rValue,
		pLabel+" "+pValue,
		"",
	)
}

// RenderVersionHeader returns the styled version selection
// header block.
func RenderVersionHeader() string {
	return "\n" + headingStyle.Render(
		"Gemara Schema Version Selection",
	) + "\n"
}

// RenderVersionOption formats a version option line.
func RenderVersionOption(
	label string,
	tag string,
	detail string,
) string {
	line := successStyle.Render(label + ": " + tag)
	if detail != "" {
		line += " " + faintStyle.Render("— "+detail)
	}
	return line
}

// RenderNote formats a note message.
func RenderNote(msg string) string {
	prefix := subtleStyle.Render("Note:")
	return prefix + " " + faintStyle.Render(msg)
}

// RenderWarning formats a warning message.
func RenderWarning(msg string) string {
	prefix := warningStyle.Render("Warning:")
	return prefix + " " + msg
}

// RenderSuccess formats a success message.
func RenderSuccess(msg string) string {
	return successStyle.Render("✓") + " " + msg
}

// RenderStatus formats a status/progress message.
func RenderStatus(msg string) string {
	return subtleStyle.Render("▸") + " " + msg
}

// RenderQuestion formats a demo-mode question display.
func RenderQuestion(question string) string {
	return questionStyle.Render("? " + question)
}

// RenderAnswer formats a demo-mode answer display.
func RenderAnswer(answer string) string {
	return "  " + answerStyle.Render("› "+answer)
}

// RenderLearningPath displays a styled learning path with
// numbered steps, left-bar accent, layer badges, and
// color-coded why/how/what annotations.
func RenderLearningPath(
	path *tutorials.LearningPath,
	out io.Writer,
) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)

	header := headingStyle.Render(
		"Your Tailored Learning Path",
	)
	if path.TargetRole != "" {
		header += " " + roleInfoStyle.Render(
			"("+path.TargetRole+")",
		)
	}
	fmt.Fprintln(out, header)
	fmt.Fprintln(out)

	for i, step := range path.Steps {
		renderPathStep(i, step, out)
	}
}

// renderPathStep renders a single learning path step with
// a left-side color bar accent and indented annotations.
func renderPathStep(
	index int,
	step tutorials.PathStep,
	out io.Writer,
) {
	stepNum := stepNumStyle.Render(
		fmt.Sprintf("Step %d", index+1),
	)

	layerName := LayerNames[step.Layer]
	if layerName == "" {
		layerName = fmt.Sprintf("Layer %d", step.Layer)
	}
	badge := layerBadgeStyle.Render(
		fmt.Sprintf(" L%d: %s ", step.Layer, layerName),
	)

	title := tutorialTitleStyle.Render(
		step.Tutorial.Title,
	)

	whyLabel := annotationLabelStyle.Render("Why: ")
	whyText := annotationTextStyle.Render(
		step.WhyAnnotation,
	)

	howLabel := annotationLabelStyle.Render("How: ")
	howText := annotationTextStyle.Render(
		step.HowAnnotation,
	)

	whatLabel := annotationLabelStyle.Render("What:")
	whatText := annotationTextStyle.Render(
		" " + step.WhatAnnotation,
	)

	lines := []string{
		stepNum + "  " + badge + "  " + title,
		"",
		whyLabel + whyText,
		howLabel + howText,
		whatLabel + whatText,
	}

	if step.VersionMismatch != nil {
		lines = append(lines, "")
		lines = append(lines,
			RenderWarning(fmt.Sprintf(
				"Tutorial uses schema %s "+
					"(you selected %s)",
				step.VersionMismatch.TutorialVersion,
				step.VersionMismatch.SelectedVersion,
			)),
		)
	}

	content := strings.Join(lines, "\n")
	fmt.Fprintln(out, stepBarStyle.Render(content))
}

// RenderLayerBadge returns a styled inline layer badge.
func RenderLayerBadge(layer int) string {
	name := LayerNames[layer]
	if name == "" {
		name = fmt.Sprintf("Layer %d", layer)
	}
	return layerBadgeStyle.Render(
		fmt.Sprintf(" L%d ", layer),
	) + " " + faintStyle.Render(name)
}

// RenderConfidence returns a styled confidence indicator.
func RenderConfidence(strong bool) string {
	if strong {
		return confidenceStrongStyle.Render("strong")
	}
	return confidenceInferredStyle.Render("inferred")
}

// RenderKeywordTags returns styled keyword tags.
func RenderKeywordTags(keywords []string) string {
	if len(keywords) == 0 {
		return ""
	}
	var tags []string
	for _, kw := range keywords {
		tags = append(tags,
			keywordTagStyle.Render(kw),
		)
	}
	return strings.Join(tags, faintStyle.Render(", "))
}

// categoryBadgeStyle is a compact badge for block categories.
var categoryBadgeStyle = lipgloss.NewStyle().
	Foreground(colorWhite).
	Background(colorCyan).
	Padding(0, 1)

// driftAddedStyle styles added drift indicators.
var driftAddedStyle = lipgloss.NewStyle().
	Foreground(colorSuccess).
	Bold(true)

// driftModifiedStyle styles modified drift indicators.
var driftModifiedStyle = lipgloss.NewStyle().
	Foreground(colorWarning).
	Bold(true)

// driftRemovedStyle styles removed drift indicators.
var driftRemovedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FF4444")).
	Bold(true)

// RenderContentBlock displays a styled content block card
// with left-bar accent, category badge, layer badge, and
// source info.
func RenderContentBlock(
	id string,
	category string,
	layer int,
	source string,
	schemaVersion string,
	body string,
) string {
	catBadge := categoryBadgeStyle.Render(
		" " + category + " ",
	)

	layerName := LayerNames[layer]
	if layerName == "" {
		layerName = fmt.Sprintf("Layer %d", layer)
	}
	lBadge := layerBadgeStyle.Render(
		fmt.Sprintf(" L%d ", layer),
	)

	titleLine := catBadge + "  " + lBadge + "  " +
		tutorialTitleStyle.Render(id)

	srcLine := faintStyle.Render(
		"Source: " + source +
			" (" + schemaVersion + ")",
	)

	// Truncate body for display (first 3 lines).
	bodyLines := strings.SplitN(body, "\n", 4)
	preview := strings.Join(bodyLines[:min(
		len(bodyLines), 3,
	)], "\n")
	if len(bodyLines) > 3 {
		preview += "\n" + faintStyle.Render("...")
	}
	bodyText := subtleStyle.Render(preview)

	content := strings.Join([]string{
		titleLine,
		srcLine,
		"",
		bodyText,
	}, "\n")

	return stepBarStyle.Render(content)
}

// RenderDriftResult displays a drift indicator with styled
// type and block ID.
func RenderDriftResult(
	blockID string,
	driftType string,
) string {
	var indicator string
	switch driftType {
	case "added":
		indicator = driftAddedStyle.Render("+ ADDED")
	case "modified":
		indicator = driftModifiedStyle.Render(
			"~ MODIFIED",
		)
	case "removed":
		indicator = driftRemovedStyle.Render("- REMOVED")
	default:
		indicator = faintStyle.Render("? " + driftType)
	}

	return fmt.Sprintf(
		"  %s  %s",
		indicator,
		faintStyle.Render(blockID),
	)
}

// RenderBlockSummary displays a summary of extracted blocks
// by category and layer.
func RenderBlockSummary(
	total int,
	byCat map[string]int,
	byLayer map[int]int,
	out io.Writer,
) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)

	header := headingStyle.Render(
		"Content Block Extraction Summary",
	)
	fmt.Fprintln(out, header)
	fmt.Fprintln(out)

	fmt.Fprintf(out, "  %s %s\n",
		faintStyle.Render("Total blocks:"),
		successStyle.Render(
			fmt.Sprintf("%d", total),
		),
	)
	fmt.Fprintln(out)

	if len(byCat) > 0 {
		fmt.Fprintln(out,
			faintStyle.Render("  By category:"),
		)
		for cat, count := range byCat {
			fmt.Fprintf(out, "    %s %s\n",
				keywordTagStyle.Render(cat),
				faintStyle.Render(
					fmt.Sprintf("(%d)", count),
				),
			)
		}
		fmt.Fprintln(out)
	}

	if len(byLayer) > 0 {
		fmt.Fprintln(out,
			faintStyle.Render("  By layer:"),
		)
		for layer, count := range byLayer {
			fmt.Fprintf(out, "    %s %s\n",
				RenderLayerBadge(layer),
				faintStyle.Render(
					fmt.Sprintf("(%d)", count),
				),
			)
		}
		fmt.Fprintln(out)
	}
}

// handoffBarStyle uses an orange left-side accent for
// handoff point cards.
var handoffBarStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.ThickBorder()).
	BorderLeft(true).
	BorderRight(false).
	BorderTop(false).
	BorderBottom(false).
	BorderForeground(colorOrange).
	PaddingLeft(2).
	MarginLeft(2).
	MarginBottom(1)

// gapBarStyle uses a warning-colored left-side accent for
// coverage gap cards.
var gapBarStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.ThickBorder()).
	BorderLeft(true).
	BorderRight(false).
	BorderTop(false).
	BorderBottom(false).
	BorderForeground(colorWarning).
	PaddingLeft(2).
	MarginLeft(2).
	MarginBottom(1)

// RenderCollaborationView displays a styled team
// collaboration view with member-layer grid, handoff
// points, and coverage gaps.
func RenderCollaborationView(
	view *team.CollaborationView,
	out io.Writer,
) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, RenderDivider())
	fmt.Fprintln(out)

	header := headingStyle.Render(
		"Team Collaboration View",
	)
	header += " " + roleInfoStyle.Render(
		"("+view.TeamName+")",
	)
	fmt.Fprintln(out, header)
	fmt.Fprintln(out)

	// Member-layer grid using table.
	renderMemberGrid(view, out)

	// Handoff points.
	if len(view.Handoffs) > 0 {
		fmt.Fprintln(out, headingStyle.Render(
			"Handoff Points",
		))
		fmt.Fprintln(out)

		for i, hp := range view.Handoffs {
			fmt.Fprintf(out, "%s\n",
				renderHandoffCard(i, hp),
			)
		}
	}

	// Coverage gaps.
	if len(view.CoverageGaps) > 0 {
		RenderCoverageGaps(view.CoverageGaps, out)
	}
}

// renderMemberGrid renders a table of team members and
// their Gemara layer assignments.
func renderMemberGrid(
	view *team.CollaborationView,
	out io.Writer,
) {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(
			lipgloss.NewStyle().
				Foreground(colorSubtle),
		).
		StyleFunc(
			func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return headerStyle
				}
				if col == 0 {
					return lipgloss.NewStyle().
						Bold(true).
						Foreground(colorOrange)
				}
				if col == 1 {
					return lipgloss.NewStyle().
						Foreground(colorWhite)
				}
				return faintStyle
			},
		).
		Headers("Member", "Role", "Layers")

	for _, m := range view.Members {
		var layerStrs []string
		for _, l := range m.Layers {
			name := LayerNames[l]
			if name == "" {
				name = fmt.Sprintf("Layer %d", l)
			}
			layerStrs = append(layerStrs,
				fmt.Sprintf("L%d (%s)", l, name),
			)
		}
		t.Row(
			m.Name,
			m.RoleName,
			strings.Join(layerStrs, ", "),
		)
	}

	fmt.Fprintln(out, t.String())
	fmt.Fprintln(out)
}

// renderHandoffCard renders a single handoff point card
// with an orange left-bar accent.
func renderHandoffCard(
	index int,
	hp team.HandoffPoint,
) string {
	num := stepNumStyle.Render(
		fmt.Sprintf("Handoff %d", index+1),
	)

	producerBadge := layerBadgeStyle.Render(
		fmt.Sprintf(" L%d ", hp.ProducerLayer),
	)
	consumerBadge := layerBadgeStyle.Render(
		fmt.Sprintf(" L%d ", hp.ConsumerLayer),
	)

	flow := roleInfoStyle.Render(hp.ProducerName) +
		" " + producerBadge +
		faintStyle.Render(" → ") +
		roleInfoStyle.Render(hp.ConsumerName) +
		" " + consumerBadge

	lines := []string{
		num + "  " + flow,
	}

	if hp.Description != "" {
		lines = append(lines, "")
		lines = append(lines,
			faintStyle.Render(hp.Description),
		)
	}

	if len(hp.ArtifactTypes) > 0 {
		lines = append(lines, "")
		label := annotationLabelStyle.Render(
			"Artifacts: ",
		)
		arts := make([]string, len(hp.ArtifactTypes))
		for i, at := range hp.ArtifactTypes {
			arts[i] = keywordTagStyle.Render(at)
		}
		lines = append(lines,
			label+strings.Join(
				arts, faintStyle.Render(", "),
			),
		)
	}

	if len(hp.ProducerTutorials) > 0 {
		lines = append(lines, "")
		label := annotationLabelStyle.Render(
			"Producer tutorials: ",
		)
		lines = append(lines,
			label+faintStyle.Render(
				strings.Join(
					hp.ProducerTutorials, ", ",
				),
			),
		)
	}

	if len(hp.ConsumerTutorials) > 0 {
		label := annotationLabelStyle.Render(
			"Consumer tutorials: ",
		)
		lines = append(lines,
			label+faintStyle.Render(
				strings.Join(
					hp.ConsumerTutorials, ", ",
				),
			),
		)
	}

	content := strings.Join(lines, "\n")
	return handoffBarStyle.Render(content)
}

// RenderHandoffPoint returns a styled handoff point card
// for detailed inspection.
func RenderHandoffPoint(hp team.HandoffPoint) string {
	return renderHandoffCard(0, hp)
}

// RenderCoverageGaps displays warnings for Gemara layers
// with no assigned team member.
func RenderCoverageGaps(gaps []int, out io.Writer) {
	fmt.Fprintln(out, headingStyle.Render(
		"Coverage Gaps",
	))
	fmt.Fprintln(out)

	var lines []string
	for _, l := range gaps {
		name := LayerNames[l]
		if name == "" {
			name = fmt.Sprintf("Layer %d", l)
		}
		lines = append(lines, fmt.Sprintf(
			"%s %s — no team member assigned",
			warningStyle.Render("▪"),
			fmt.Sprintf("L%d (%s)", l, name),
		))
	}

	content := strings.Join(lines, "\n")
	fmt.Fprintln(out, gapBarStyle.Render(content))
}

// RenderTeamMember returns a styled team member card.
func RenderTeamMember(
	name string,
	roleName string,
	layers []int,
) string {
	memberName := roleInfoStyle.Render(name)
	role := faintStyle.Render("(" + roleName + ")")

	var badges []string
	for _, l := range layers {
		badges = append(badges, RenderLayerBadge(l))
	}

	lines := []string{
		memberName + " " + role,
		"",
		strings.Join(badges, "  "),
	}

	content := strings.Join(lines, "\n")
	return stepBarStyle.Render(content)
}

// RenderAuthoringStep returns a styled authoring step card
// with progress indicator, field list, and role-specific
// explanation.
func RenderAuthoringStep(
	stepName string,
	description string,
	roleExplanation string,
	fieldNames []string,
	current int,
	total int,
) string {
	progress := fmt.Sprintf(
		"Step %d of %d", current+1, total,
	)
	header := stepNumStyle.Render(progress) + "  " +
		headingStyle.Render(stepName)

	var lines []string
	lines = append(lines, header)
	lines = append(lines, "")
	lines = append(lines,
		subtleStyle.Render(description),
	)

	if roleExplanation != "" {
		lines = append(lines, "")
		lines = append(lines,
			annotationLabelStyle.Render(
				"Why: ",
			)+annotationTextStyle.Render(
				roleExplanation,
			),
		)
	}

	if len(fieldNames) > 0 {
		lines = append(lines, "")
		lines = append(lines,
			faintStyle.Render("Fields:"),
		)
		for _, f := range fieldNames {
			lines = append(lines,
				"  "+subtleStyle.Render("- ")+f,
			)
		}
	}

	content := strings.Join(lines, "\n")
	return stepBarStyle.Render(content)
}

// RenderFieldPrompt returns a styled field input prompt
// with description and example value.
func RenderFieldPrompt(
	name string,
	description string,
	exampleValue string,
	required bool,
) string {
	label := headingStyle.Render(name)
	if required {
		label += warningStyle.Render(" *")
	}
	desc := subtleStyle.Render(description)

	result := label + "\n" + desc
	if exampleValue != "" {
		result += "\n" + faintStyle.Render(
			"Example: "+exampleValue,
		)
	}
	return result
}

// RenderValidationResults returns a styled display of
// validation errors with fix suggestions.
func RenderValidationResults(
	errs []validationDisplayError,
) string {
	if len(errs) == 0 {
		return successStyle.Render(
			"  Validation passed",
		)
	}

	var lines []string
	lines = append(lines,
		warningStyle.Render(fmt.Sprintf(
			"  %d validation error(s):",
			len(errs),
		)),
	)
	for _, e := range errs {
		lines = append(lines, "")
		if e.FieldPath != "" {
			lines = append(lines,
				"  "+warningStyle.Render("Field: ")+
					e.FieldPath,
			)
		}
		lines = append(lines,
			"  "+subtleStyle.Render("Error: ")+
				e.Message,
		)
		if e.FixSuggestion != "" {
			lines = append(lines,
				"  "+successStyle.Render("Fix: ")+
					e.FixSuggestion,
			)
		}
	}
	return strings.Join(lines, "\n")
}

// validationDisplayError is a display-only struct for
// rendering validation errors without importing the
// authoring package.
type validationDisplayError struct {
	FieldPath     string
	Message       string
	FixSuggestion string
}

// RenderArtifactSummary returns a styled overview of a
// completed authored artifact.
func RenderArtifactSummary(
	artifactType string,
	schemaDef string,
	schemaVersion string,
	sectionCount int,
	status string,
) string {
	var lines []string
	lines = append(lines,
		headingStyle.Render("Authored Artifact"),
	)
	lines = append(lines, "")
	lines = append(lines,
		annotationLabelStyle.Render("Type: ")+
			artifactType,
	)
	lines = append(lines,
		annotationLabelStyle.Render("Schema: ")+
			schemaDef+" ("+schemaVersion+")",
	)
	lines = append(lines,
		annotationLabelStyle.Render("Sections: ")+
			fmt.Sprintf("%d", sectionCount),
	)

	statusStr := status
	statusRender := faintStyle.Render(statusStr)
	if status == "valid" {
		statusRender = successStyle.Render(statusStr)
	} else if status == "invalid" {
		statusRender = warningStyle.Render(statusStr)
	}
	lines = append(lines,
		annotationLabelStyle.Render("Status: ")+
			statusRender,
	)

	content := strings.Join(lines, "\n")
	return stepBarStyle.Render(content)
}

// RenderAuthoringProgress returns a styled progress
// indicator for the authoring flow.
func RenderAuthoringProgress(
	completed int,
	total int,
) string {
	pct := 0
	if total > 0 {
		pct = (completed * 100) / total
	}

	barWidth := 20
	filled := (pct * barWidth) / 100
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) +
		strings.Repeat("░", barWidth-filled)

	return fmt.Sprintf(
		"%s %s %d%%",
		faintStyle.Render("Progress:"),
		successStyle.Render(bar),
		pct,
	)
}

// RenderTutorialHeader renders the header for a tutorial
// step with title, layer badge, role-specific motivation,
// and section count.
func RenderTutorialHeader(
	title string,
	layer int,
	whyAnnotation string,
	sectionCount int,
) string {
	var lines []string
	lines = append(lines,
		tutorialTitleStyle.Render(title)+
			"  "+RenderLayerBadge(layer),
	)
	lines = append(lines, "")
	if whyAnnotation != "" {
		lines = append(lines,
			annotationLabelStyle.Render("Why: ")+
				annotationTextStyle.Render(
					whyAnnotation,
				),
		)
		lines = append(lines, "")
	}
	lines = append(lines,
		faintStyle.Render(fmt.Sprintf(
			"%d sections to complete", sectionCount,
		)),
	)
	content := strings.Join(lines, "\n")
	return stepBarStyle.Render(content)
}

// RenderTutorialSection renders a single tutorial section
// with its heading, body content, position indicator, and
// role-personalized context.
func RenderTutorialSection(
	heading string,
	body string,
	current int,
	total int,
	tutorialTitle string,
	roleName string,
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

	// Render the section body with wrapping.
	if body != "" {
		bodyLines := strings.Split(
			strings.TrimSpace(body), "\n",
		)
		for _, bl := range bodyLines {
			bl = strings.TrimSpace(bl)
			if bl == "" {
				lines = append(lines, "")
				continue
			}
			lines = append(lines, bl)
		}
	} else {
		lines = append(lines,
			faintStyle.Render(
				"(This section has no content yet.)",
			),
		)
	}

	// Add role context if available.
	if roleName != "" {
		lines = append(lines, "")
		lines = append(lines,
			annotationLabelStyle.Render("Context: ")+
				annotationTextStyle.Render(
					fmt.Sprintf(
						"As a %s, apply this "+
							"section's concepts to "+
							"your daily workflows.",
						roleName,
					),
				),
		)
	}

	content := strings.Join(lines, "\n")
	return stepBarStyle.Render(content)
}

// min returns the smaller of two ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
