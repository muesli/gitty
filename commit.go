package main

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/shurcooL/githubv4"
)

var historyQuery struct {
	Repository struct {
		Object struct {
			Commit struct {
				Oid     githubv4.String
				History struct {
					TotalCount githubv4.Int
					Edges      []struct {
						Cursor githubv4.String
						Node   struct {
							QLCommit
						}
					}
				} `graphql:"history(first: 100, since: $since)"`
			} `graphql:"... on Commit"`
		} `graphql:"object(expression: \"HEAD\")"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type QLCommit struct {
	OID             githubv4.GitObjectID
	MessageHeadline githubv4.String
	CommittedDate   githubv4.GitTimestamp
	Author          struct {
		User struct {
			Login githubv4.String
		}
	}
}

type Commit struct {
	ID              string
	MessageHeadline string
	CommittedAt     time.Time
	Author          string
}

func history(owner string, name string, since time.Time) ([]Commit, error) {
	var commits []Commit

	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
		"since": githubv4.GitTimestamp{Time: since},
	}

	// if err := client.Query(context.Background(), &historyQuery, variables); err != nil {
	if err := queryWithRetry(context.Background(), &historyQuery, variables); err != nil {
		return commits, err
	}

	for _, v := range historyQuery.Repository.Object.Commit.History.Edges {
		if v.Node.QLCommit.OID == "" {
			// fmt.Println("Commit ID broken:", v.Node.QLCommit.OID)
			continue
		}
		commits = append(commits, CommitFromQL(v.Node.QLCommit))
	}

	return commits, nil
}

func CommitFromQL(commit QLCommit) Commit {
	return Commit{
		ID:              string(commit.OID),
		MessageHeadline: string(commit.MessageHeadline),
		CommittedAt:     commit.CommittedDate.Time,
		Author:          string(commit.Author.User.Login),
	}
}

func printCommit(commit Commit) {
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
	s += titleStyle.Render(truncate.String(commit.MessageHeadline, 80-7))
	s += genericStyle.Render(" ")
	s += timeStyle.Render(ago(commit.CommittedAt))
	s += genericStyle.Render(" ")
	s += numberStyle.Render(commit.Author)

	fmt.Println(s)
}

func printCommits(repo Repo) {
	commits := repo.LastRelease.CommitsSince

	headerStyle := lipgloss.NewStyle().
		PaddingTop(1).
		Foreground(lipgloss.Color(theme.colorMagenta))

	sinceTag := repo.LastRelease.TagName
	if sinceTag == "" {
		sinceTag = "creation"
	}
	if len(commits) == 0 {
		fmt.Println(headerStyle.Render(fmt.Sprintf("🔥 No new commits since %s", sinceTag)))
		return
	}
	fmt.Println(headerStyle.Render(fmt.Sprintf("🔥 %d commits since %s", len(commits), sinceTag)))

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
