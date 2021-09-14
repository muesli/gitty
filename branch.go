package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/gitty/vcs"
	"github.com/muesli/reflow/truncate"
)

func printBranch(branch vcs.Branch, maxWidth int) {
	genericStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGray))
	numberStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorBlue)).Width(maxWidth)
	authorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorBlue))
	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGreen)).Width(8).Align(lipgloss.Right)
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorDarkGray)).Width(80 - maxWidth)

	var s string
	s += numberStyle.Render(branch.Name)
	s += genericStyle.Render(" ")
	s += titleStyle.Render(truncate.StringWithTail(branch.LastCommit.MessageHeadline, uint(80-maxWidth), "…"))
	s += genericStyle.Render(" ")
	s += timeStyle.Render(ago(branch.LastCommit.CommittedAt))
	s += genericStyle.Render(" ")
	s += authorStyle.Render(branch.LastCommit.Author)

	fmt.Println(s)
}

func printBranches(branches []vcs.Branch) {
	headerStyle := lipgloss.NewStyle().
		PaddingTop(1).
		Foreground(lipgloss.Color(theme.colorMagenta))

	sort.Slice(branches, func(i, j int) bool {
		if branches[i].LastCommit.CommittedAt.Equal(branches[j].LastCommit.CommittedAt) {
			return strings.Compare(branches[i].Name, branches[j].Name) < 0
		}
		return branches[i].LastCommit.CommittedAt.After(branches[j].LastCommit.CommittedAt)
	})

	// filter list
	var b []vcs.Branch
	for _, v := range branches {
		if *maxBranchAge > 0 &&
			v.LastCommit.CommittedAt.Before(time.Now().Add(-24*time.Duration(*maxBranchAge)*time.Hour)) {
			continue
		}
		b = append(b, v)
	}
	branches = b

	// trimmed := false
	if *maxBranches > 0 && len(branches) > *maxBranches {
		branches = branches[:*maxBranches]
		// trimmed = true
	}

	fmt.Println(headerStyle.Render(fmt.Sprintf("%s %s", "🌳", pluralize(len(branches), "active branch", "active branches"))))

	// detect max width of branch name
	var maxWidth int
	for _, v := range branches {
		if len(v.Name) > maxWidth {
			maxWidth = len(v.Name)
		}
	}

	for _, v := range branches {
		printBranch(v, maxWidth)
	}
	// if trimmed {
	// 	fmt.Println("...")
	// }
}
