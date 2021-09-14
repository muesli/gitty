package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/muesli/gitty/vcs"
	"github.com/muesli/reflow/truncate"
)

func printCommit(commit vcs.Commit) {
	genericStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGray))
	numberStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorBlue))
	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGreen)).Width(8).Align(lipgloss.Right)
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorDarkGray)).Width(80 - 7)

	var s string
	s += numberStyle.Render(commit.ID[:7])
	s += genericStyle.Render(" ")
	s += titleStyle.Render(truncate.StringWithTail(commit.MessageHeadline, 80-7, "…"))
	s += genericStyle.Render(" ")
	s += timeStyle.Render(ago(commit.CommittedAt))
	s += genericStyle.Render(" ")
	s += numberStyle.Render(commit.Author)

	fmt.Println(s)
}

func printCommits(repo vcs.Repo) {
	commits := repo.LastRelease.CommitsSince

	// dimColor := gamut.ToHex(gamut.Darker(gamut.Hex(theme.colorMagenta), 0.40))

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorMagenta))
	// headerDimStyle := lipgloss.NewStyle().
	// 	Foreground(lipgloss.Color(dimColor))
	sinceTag := repo.LastRelease.TagName
	if sinceTag == "" {
		sinceTag = "creation"
	}

	fmt.Printf("\n🔥 %s %s\n",
		headerStyle.Render(fmt.Sprintf("%s %s",
			pluralize(len(commits), "commit since", "commits since"),
			sinceTag)),

		headerStyle.Render(fmt.Sprintf("(%s)",
			humanize.Time(repo.LastRelease.PublishedAt))),
	)

	// trimmed := false
	if *maxCommits > 0 && len(commits) > *maxCommits {
		commits = commits[:*maxCommits]
		// trimmed = true
	}

	for _, v := range commits {
		printCommit(v)
	}
	// if trimmed {
	// 	fmt.Println("...")
	// }
}
