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

	panelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	// Additional styles for the role discovery flow.
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

	stepCardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Padding(0, 2).
			MarginLeft(2)

	roleInfoStyle = lipgloss.NewStyle().
			Foreground(colorOrange).
			Bold(true)
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

// RenderSessionStatus returns a styled session summary
// including role and learning path information.
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

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		versionLine,
		fallbackLine,
	)

	return "\n" + panelStyle.Render(content) + "\n"
}

// RenderSessionRoleInfo returns a styled role/path summary
// to display after the session panel.
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
// numbered steps inside bordered cards, layer badges, and
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

// renderPathStep renders a single learning path step as a
// styled card with layer badge and annotations.
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

	// Build the card content.
	whyLabel := annotationLabelStyle.Render("Why:")
	whyText := annotationTextStyle.Render(
		" " + step.WhyAnnotation,
	)

	howLabel := annotationLabelStyle.Render("How:")
	howText := annotationTextStyle.Render(
		" " + step.HowAnnotation,
	)

	whatLabel := annotationLabelStyle.Render("What:")
	whatText := annotationTextStyle.Render(
		" " + step.WhatAnnotation,
	)

	var cardLines []string
	cardLines = append(cardLines, "")
	cardLines = append(cardLines,
		stepNum+"  "+badge+"  "+title,
	)
	cardLines = append(cardLines, "")
	cardLines = append(cardLines,
		"  "+whyLabel+whyText,
	)
	cardLines = append(cardLines,
		"  "+howLabel+howText,
	)
	cardLines = append(cardLines,
		"  "+whatLabel+whatText,
	)

	if step.VersionMismatch != nil {
		cardLines = append(cardLines, "")
		cardLines = append(cardLines,
			"  "+RenderWarning(fmt.Sprintf(
				"Tutorial uses schema %s "+
					"(you selected %s)",
				step.VersionMismatch.TutorialVersion,
				step.VersionMismatch.SelectedVersion,
			)),
		)
	}

	cardLines = append(cardLines, "")

	cardContent := strings.Join(cardLines, "\n")
	fmt.Fprintln(out, stepCardStyle.Render(cardContent))
	fmt.Fprintln(out)
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
