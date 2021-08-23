package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/shurcooL/githubv4"
)

var recentReleasesQuery struct {
	User struct {
		Login                     githubv4.String
		RepositoriesContributedTo struct {
			TotalCount githubv4.Int
			Edges      []struct {
				Cursor githubv4.String
				Node   struct {
					QLRepository
					Releases QLRelease `graphql:"releases(first: 10, orderBy: {field: CREATED_AT, direction: DESC})"`
				}
			}
		} `graphql:"repositoriesContributedTo(first: 100, after:$after includeUserRepositories: true, contributionTypes: COMMIT)"`
	} `graphql:"user(login:$username)"`
}

type QLRelease struct {
	Nodes []struct {
		Name         githubv4.String
		TagName      githubv4.String
		PublishedAt  githubv4.DateTime
		URL          githubv4.String
		IsPrerelease githubv4.Boolean
		IsDraft      githubv4.Boolean
	}
}

type Release struct {
	Name         string
	TagName      string
	PublishedAt  time.Time
	URL          string
	CommitsSince []Commit
}

func repoRelease(repo Repo) {
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
	commitStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorDarkGray))

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
	s += repoStyle.Render(repo.Name)
	s += versionStyle.Render(" " + repo.LastRelease.TagName)
	s += genericStyle.Render(" (")
	s += dateStyle.Render(humanize.Time(repo.LastRelease.PublishedAt))
	s += genericStyle.Render(", ")
	s += changesStyle.Render(fmt.Sprintf("%d new commits since", len(repo.LastRelease.CommitsSince)))
	s += genericStyle.Render(")")

	if *withCommits && len(repo.LastRelease.CommitsSince) > 0 {
		s += "\n"
		for i, commit := range repo.LastRelease.CommitsSince {
			if i >= *maxCommits && *maxCommits > 0 {
				break
			}

			s += commitStyle.Render(commit.ID[0:7]) + " "
			s += commitStyle.Render(commit.MessageHeadline) + "\n"
		}
	}

	fmt.Println(s)
}

func reposWithRelease(repos []Repo) []Repo {
	var r []Repo

	for _, repo := range repos {
		if repo.LastRelease.PublishedAt.IsZero() {
			continue
		}

		r = append(r, repo)
	}

	return r
}

func ReleaseFromQL(release QLRelease) Release {
	return Release{
		Name:        string(release.Nodes[0].Name),
		TagName:     string(release.Nodes[0].TagName),
		PublishedAt: release.Nodes[0].PublishedAt.Time,
		URL:         string(release.Nodes[0].URL),
	}
}
