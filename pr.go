package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/gitty/vcs"
	"github.com/muesli/reflow/truncate"
)

func printPullRequest(pr vcs.PullRequest, maxWidth int) {
	genericStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGray))
	numberStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorBlue)).Width(maxWidth).Align(lipgloss.Right)
	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGreen)).Width(8).Align(lipgloss.Right)
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorDarkGray)).Width(80 - maxWidth)

	var s string
	s += numberStyle.Render(strconv.Itoa(pr.ID))
	s += genericStyle.Render(" ")
	s += titleStyle.Render(truncate.StringWithTail(pr.Title, uint(80-maxWidth), "…"))
	s += genericStyle.Render(" ")
	s += timeStyle.Render(ago(pr.CreatedAt))
	s += genericStyle.Render(" ")
	s += pr.Labels.View()

	fmt.Println(s)
}

func printPullRequests(prs []vcs.PullRequest) {
	headerStyle := lipgloss.NewStyle().
		PaddingTop(1).
		Foreground(lipgloss.Color(theme.colorMagenta))

	fmt.Println(headerStyle.Render(fmt.Sprintf("%s %s", "📌", pluralize(len(prs), "open pull request", "open pull requests"))))

	// trimmed := false
	if *maxPullRequests > 0 && len(prs) > *maxPullRequests {
		prs = prs[:*maxPullRequests]
		// trimmed = true
	}

	// detect max width of pr number
	var maxWidth int
	for _, v := range prs {
		if len(strconv.Itoa(v.ID)) > maxWidth {
			maxWidth = len(strconv.Itoa(v.ID))
		}
	}

	for _, v := range prs {
		printPullRequest(v, maxWidth)
	}
	// if trimmed {
	// 	fmt.Println("...")
	// }
}
