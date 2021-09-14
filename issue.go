package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/gitty/vcs"
	"github.com/muesli/reflow/truncate"
)

func printIssue(issue vcs.Issue, maxWidth int) {
	genericStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGray))
	numberStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorBlue)).Width(maxWidth).Align(lipgloss.Right)
	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGreen)).Width(8).Align(lipgloss.Right)
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorDarkGray)).Width(80 - maxWidth)

	var s string
	s += numberStyle.Render(strconv.Itoa(issue.ID))
	s += genericStyle.Render(" ")
	s += titleStyle.Render(truncate.StringWithTail(issue.Title, uint(80-maxWidth), "…"))
	s += genericStyle.Render(" ")
	s += timeStyle.Render(ago(issue.CreatedAt))
	s += genericStyle.Render(" ")
	s += issue.Labels.View()

	fmt.Println(s)
}

func printIssues(issues []vcs.Issue) {
	headerStyle := lipgloss.NewStyle().
		PaddingTop(1).
		Foreground(lipgloss.Color(theme.colorMagenta))

	fmt.Println(headerStyle.Render(fmt.Sprintf("%s %s", "🐛", pluralize(len(issues), "open issue", "open issues"))))

	// trimmed := false
	if *maxIssues > 0 && len(issues) > *maxIssues {
		issues = issues[:*maxIssues]
		// trimmed = true
	}

	// detect max width of issue number
	var maxWidth int
	for _, v := range issues {
		if len(strconv.Itoa(v.ID)) > maxWidth {
			maxWidth = len(strconv.Itoa(v.ID))
		}
	}

	for _, v := range issues {
		printIssue(v, maxWidth)
	}
	// if trimmed {
	// 	fmt.Println("...")
	// }
}
