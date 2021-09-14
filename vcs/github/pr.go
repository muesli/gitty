package github

import (
	"context"

	"github.com/muesli/gitty/vcs"
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

func (c *Client) PullRequests(owner string, name string) ([]vcs.PullRequest, error) {
	var after *githubv4.String
	var pullRequests []vcs.PullRequest

	for {
		variables := map[string]interface{}{
			"owner": githubv4.String(owner),
			"name":  githubv4.String(name),
			"after": after,
		}

		if err := c.queryWithRetry(context.Background(), &pullRequestQuery, variables); err != nil {
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

func PullRequestFromQL(pr QLPullRequest) vcs.PullRequest {
	p := vcs.PullRequest{
		ID:        int(pr.Number),
		Body:      string(pr.Body),
		Title:     string(pr.Title),
		CreatedAt: pr.CreatedAt.Time,
	}

	for _, v := range pr.Labels.Edges {
		p.Labels = append(p.Labels, vcs.Label{
			Name:  string(v.Node.Name),
			Color: "#" + string(v.Node.Color),
		})
	}

	return p
}
