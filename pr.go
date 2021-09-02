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

var pullRequestQuery struct {
	Repository struct {
		PullRequests struct {
			TotalCount githubv4.Int
			Edges      []struct {
				Cursor githubv4.String
				Node   struct {
					QLPullRequest
				}
			}
		} `graphql:"pullRequests(first: 100, after: $after, states: OPEN, orderBy: {field: CREATED_AT, direction: DESC})"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type QLPullRequest struct {
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

type PullRequest struct {
	ID        int
	Body      string
	Title     string
	Labels    []Label
	CreatedAt time.Time
}

func pullRequests(owner string, name string) ([]PullRequest, error) {
	var after *githubv4.String
	var pullRequests []PullRequest

	for {
		variables := map[string]interface{}{
			"owner": githubv4.String(owner),
			"name":  githubv4.String(name),
			"after": after,
		}

		if err := client.Query(context.Background(), &pullRequestQuery, variables); err != nil {
			// if err := queryWithRetry(context.Background(), &pullRequestQuery, variables); err != nil {
			return pullRequests, err
		}
		if len(pullRequestQuery.Repository.PullRequests.Edges) == 0 {
			break
		}

		for _, v := range pullRequestQuery.Repository.PullRequests.Edges {
			pullRequests = append(pullRequests, PullRequestFromQL(v.Node.QLPullRequest))

			after = &v.Cursor
		}
	}

	return pullRequests, nil
}

func PullRequestFromQL(pr QLPullRequest) PullRequest {
	p := PullRequest{
		ID:        int(pr.Number),
		Body:      string(pr.Body),
		Title:     string(pr.Title),
		CreatedAt: pr.CreatedAt.Time,
	}

	for _, v := range pr.Labels.Edges {
		p.Labels = append(p.Labels, Label{string(v.Node.Name), string(v.Node.Color)})
	}

	return p
}

func printPullRequest(pr PullRequest, maxWidth int) {
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
	s += titleStyle.Render(truncate.String(pr.Title, uint(80-maxWidth)))
	s += genericStyle.Render(" ")
	s += timeStyle.Render(ago(pr.CreatedAt))
	s += genericStyle.Render(" ")

	for _, v := range pr.Labels {
		labelStyle := lipgloss.NewStyle().
			// Foreground(lipgloss.Color(theme.colorBlack)).
			Foreground(lipgloss.Color("#" + v.Color))

		s += labelStyle.Render(fmt.Sprintf("◖%s◗", v.Name))
		s += genericStyle.Render(" ")
	}

	fmt.Println(s)
}

func printPullRequests(prs []PullRequest) {
	headerStyle := lipgloss.NewStyle().
		PaddingTop(1).
		Foreground(lipgloss.Color(theme.colorMagenta))

	fmt.Println(headerStyle.Render(fmt.Sprintf("📌 %d open pull requests", len(prs))))

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
