package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/muesli/gitty/vcs"
	"github.com/muesli/termenv"
)

func repoRelease(repo vcs.Repo) {
	genericStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGray))
	repoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorBlue))
	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorMagenta))
	dateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGreen))
	changesStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGreen))

	day := time.Hour * 24
	week := day * 7
	month := week * 4
	since := time.Since(repo.LastRelease.PublishedAt)
	switch {
	case since > month*6:
		dateStyle = dateStyle.Foreground(lipgloss.Color(theme.colorRed))
	case since > month*3:
		dateStyle = dateStyle.Foreground(lipgloss.Color(theme.colorYellow))
	}

	switch {
	case len(repo.LastRelease.CommitsSince) > 32:
		changesStyle = changesStyle.Foreground(lipgloss.Color(theme.colorRed))
	case len(repo.LastRelease.CommitsSince) > 16:
		changesStyle = changesStyle.Foreground(lipgloss.Color(theme.colorYellow))
	}

	if len(repo.LastRelease.CommitsSince) < *minNewCommits {
		if *skipStaleRepos {
			return
		}
		// 	genericStyle = genericStyle.Foreground(lipgloss.Color(theme.colorGray))
		// 	repoStyle = repoStyle.Foreground(lipgloss.Color(theme.colorGray))
		// 	versionStyle = versionStyle.Foreground(lipgloss.Color(theme.colorGray))
		// 	dateStyle = dateStyle.Foreground(lipgloss.Color(theme.colorGray))
		// 	changesStyle = changesStyle.Foreground(lipgloss.Color(theme.colorGray))
	}

	var s string
	name := repoStyle.Render(repo.Name)
	if useLinks {
		name = termenv.Hyperlink(repo.URL, name)
	}
	s += name + " "
	release := versionStyle.Render(repo.LastRelease.TagName)
	if useLinks {
		release = termenv.Hyperlink(repo.LastRelease.URL, release)
	}
	s += release
	s += genericStyle.Render(" (")
	s += dateStyle.Render(humanize.Time(repo.LastRelease.PublishedAt))
	s += genericStyle.Render(", ")
	s += changesStyle.Render(fmt.Sprintf("%d new commits since", len(repo.LastRelease.CommitsSince)))
	s += genericStyle.Render(")")
	fmt.Println(s)

	if *withCommits && len(repo.LastRelease.CommitsSince) > 0 {
		for i, commit := range repo.LastRelease.CommitsSince {
			if i >= *maxCommits && *maxCommits > 0 {
				break
			}

			printCommit(commit)
		}

		fmt.Println()
	}
}
