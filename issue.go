package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/shurcooL/githubv4"
)

var issuesQuery struct {
	Repository struct {
		Issues struct {
			TotalCount githubv4.Int
			Edges      []struct {
				Cursor githubv4.String
				Node   struct {
					QLIssue
				}
			}
		} `graphql:"issues(first: 100, after: $after, states: OPEN, orderBy: {field: CREATED_AT, direction: DESC})"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type QLIssue struct {
	Number    githubv4.Int
	Body      githubv4.String
	Title     githubv4.String
	CreatedAt githubv4.DateTime
	Labels    struct {
		Edges []struct {
			Cursor githubv4.String
			Node   struct {
				Name  githubv4.String
				Color githubv4.String
			}
		}
	} `graphql:"labels(first: 100, orderBy: {field: NAME, direction: ASC})"`
}

type Issue struct {
	ID        int
	Body      string
	Title     string
	Labels    []Label
	CreatedAt time.Time
}

func issues(owner string, name string) ([]Issue, error) {
	var after *githubv4.String
	var issues []Issue

	for {
		variables := map[string]interface{}{
			"owner": githubv4.String(owner),
			"name":  githubv4.String(name),
			"after": after,
		}

		if err := queryWithRetry(context.Background(), &issuesQuery, variables); err != nil {
			return issues, err
		}
		if len(issuesQuery.Repository.Issues.Edges) == 0 {
			break
		}

		for _, v := range issuesQuery.Repository.Issues.Edges {
			issues = append(issues, IssueFromQL(v.Node.QLIssue))

			after = &v.Cursor
		}
	}

	return issues, nil
}

func IssueFromQL(issue QLIssue) Issue {
	i := Issue{
		ID:        int(issue.Number),
		Body:      string(issue.Body),
		Title:     string(issue.Title),
		CreatedAt: issue.CreatedAt.Time,
	}

	for _, v := range issue.Labels.Edges {
		i.Labels = append(i.Labels, Label{string(v.Node.Name), string(v.Node.Color)})
	}

	return i
}

func printIssue(issue Issue, maxWidth int) {
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
	s += titleStyle.Render(truncate.String(issue.Title, uint(80-maxWidth)))
	s += genericStyle.Render(" ")
	s += timeStyle.Render(ago(issue.CreatedAt))
	s += genericStyle.Render(" ")

	for _, v := range issue.Labels {
		labelStyle := lipgloss.NewStyle().
			// Foreground(lipgloss.Color(theme.colorBlack)).
			Foreground(lipgloss.Color("#" + v.Color))

		s += labelStyle.Render(fmt.Sprintf("◖%s◗", v.Name))
		s += genericStyle.Render(" ")
	}

	fmt.Println(s)
}

func printIssues(issues []Issue) {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorMagenta))

	fmt.Println(headerStyle.Render(fmt.Sprintf("🐛 %d open issues", len(issues))))

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
