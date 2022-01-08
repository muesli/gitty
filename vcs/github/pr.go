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
					qlPullRequest
				}
			}
		} `graphql:"pullRequests(first: 100, after: $after, states: OPEN, orderBy: {field: CREATED_AT, direction: DESC})"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type qlPullRequest struct {
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

// PullRequests returns a list of pull requests for the given repository.
func (c *Client) PullRequests(owner string, name string) ([]vcs.PullRequest, error) {
	var pullRequests []vcs.PullRequest

	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
		"after": (*githubv4.String)(nil),
	}

	for {
		if err := c.queryWithRetry(context.Background(), &pullRequestQuery, variables); err != nil {
			return pullRequests, err
		}
		if len(pullRequestQuery.Repository.PullRequests.Edges) == 0 {
			break
		}

		for _, v := range pullRequestQuery.Repository.PullRequests.Edges {
			pullRequests = append(pullRequests, pullRequestFromQL(v.Node.qlPullRequest))

			variables["after"] = githubv4.NewString(v.Cursor)
		}
	}

	return pullRequests, nil
}

func pullRequestFromQL(pr qlPullRequest) vcs.PullRequest {
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
