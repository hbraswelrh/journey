// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"

	lipgloss "charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/hbraswelrh/pacman/internal/consts"
	"github.com/hbraswelrh/pacman/internal/tutorials"
)

// Color palette for the Pac-Man TUI.
var (
	colorPrimary = lipgloss.Color("#7D56F4")
	colorSubtle  = lipgloss.Color("#6C6C6C")
	colorSuccess = lipgloss.Color("#04B575")
	colorWarning = lipgloss.Color("#FFCC00")
	colorFaint   = lipgloss.Color("#999999")
)

// Reusable text styles.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
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
)

// RenderBanner returns the styled application banner.
func RenderBanner() string {
	banner := titleStyle.Render(
		"  Pac-Man — Gemara Tutorial Engine  ",
	)
	return "\n" + banner + "\n"
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
						Foreground(
							lipgloss.Color(
								"#FAFAFA",
							),
						)
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
func RenderSessionStatus(
	version string,
	fallback bool,
) string {
	header := headingStyle.Render("Session Ready")

	versionLine := "  Schema version: " +
		successStyle.Render(version)

	modeLabel := successStyle.Render("false")
	if fallback {
		modeLabel = warningStyle.Render("true")
	}
	fallbackLine := "  Fallback mode:  " + modeLabel

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		versionLine,
		fallbackLine,
	)

	return "\n" + panelStyle.Render(content) + "\n"
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

// RenderLearningPath displays a styled learning path with
// numbered steps, layer badges, and why/how/what sections.
func RenderLearningPath(
	path *tutorials.LearningPath,
	out io.Writer,
) {
	fmt.Fprintln(out)
	header := headingStyle.Render(
		"Your Tailored Learning Path",
	)
	if path.TargetRole != "" {
		header += " " + faintStyle.Render(
			"("+path.TargetRole+")",
		)
	}
	fmt.Fprintln(out, header)
	fmt.Fprintln(out)

	for i, step := range path.Steps {
		stepNum := fmt.Sprintf("Step %d", i+1)
		layerBadge := fmt.Sprintf(
			"Layer %d", step.Layer,
		)

		fmt.Fprintf(out, "  %s  %s  %s\n",
			successStyle.Render(stepNum),
			titleStyle.Render(" "+layerBadge+" "),
			step.Tutorial.Title,
		)

		fmt.Fprintf(out,
			"    %s %s\n",
			headingStyle.Render("Why:"),
			step.WhyAnnotation,
		)
		fmt.Fprintf(out,
			"    %s %s\n",
			headingStyle.Render("How:"),
			step.HowAnnotation,
		)
		fmt.Fprintf(out,
			"    %s %s\n",
			headingStyle.Render("What:"),
			step.WhatAnnotation,
		)

		if step.VersionMismatch != nil {
			fmt.Fprintf(out,
				"    %s\n",
				RenderWarning(fmt.Sprintf(
					"Tutorial uses schema %s "+
						"(you selected %s)",
					step.VersionMismatch.
						TutorialVersion,
					step.VersionMismatch.
						SelectedVersion,
				)),
			)
		}

		fmt.Fprintln(out)
	}
}
