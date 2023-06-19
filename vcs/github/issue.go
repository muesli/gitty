package github

import (
	"context"

	"github.com/muesli/gitty/vcs"
	"github.com/shurcooL/githubv4"
)

var issuesQuery struct {
	Repository struct {
		Issues struct {
			TotalCount githubv4.Int
			Edges      []struct {
				Cursor githubv4.String
				Node   struct {
					qlIssue
				}
			}
		} `graphql:"issues(first: 100, after: $after, states: OPEN, orderBy: {field: CREATED_AT, direction: DESC})"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type qlIssue struct {
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
	URL githubv4.String
}

// Issues returns a list of issues for the given repository.
func (c *Client) Issues(owner string, name string) ([]vcs.Issue, error) {
	var issues []vcs.Issue

	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
		"after": (*githubv4.String)(nil),
	}

	for {
		if err := c.queryWithRetry(context.Background(), &issuesQuery, variables); err != nil {
			return issues, err
		}
		if len(issuesQuery.Repository.Issues.Edges) == 0 {
			break
		}

		for _, v := range issuesQuery.Repository.Issues.Edges {
			issues = append(issues, issueFromQL(v.Node.qlIssue))

			variables["after"] = githubv4.NewString(v.Cursor)
		}
	}

	return issues, nil
}

func issueFromQL(issue qlIssue) vcs.Issue {
	i := vcs.Issue{
		ID:        int(issue.Number),
		Body:      string(issue.Body),
		Title:     string(issue.Title),
		CreatedAt: issue.CreatedAt.Time,
		URL:       string(issue.URL),
	}

	for _, v := range issue.Labels.Edges {
		i.Labels = append(i.Labels, vcs.Label{
			Name:  string(v.Node.Name),
			Color: "#" + string(v.Node.Color),
		})
	}

	return i
}
