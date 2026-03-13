// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/tutorials"
)

// Color palette for the Pac-Man TUI.
var (
	colorPrimary   = lipgloss.Color("#7D56F4")
	colorSecondary = lipgloss.Color("#5B3CC4")
	colorSubtle    = lipgloss.Color("#6C6C6C")
	colorSuccess   = lipgloss.Color("#04B575")
	colorWarning   = lipgloss.Color("#FFCC00")
	colorFaint     = lipgloss.Color("#999999")
	colorWhite     = lipgloss.Color("#FAFAFA")
	colorDim       = lipgloss.Color("#555555")
	colorCyan      = lipgloss.Color("#00D4FF")
	colorOrange    = lipgloss.Color("#FF8C00")
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

// min returns the smaller of two ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
