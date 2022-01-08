package github

import (
	"context"

	"github.com/muesli/gitty/vcs"
	"github.com/shurcooL/githubv4"
)

var branchesQuery struct {
	Repository struct {
		Refs struct {
			Nodes []struct {
				Name   githubv4.String
				Target struct {
					Commit qlCommit `graphql:"... on Commit"`
				}
			}
		} `graphql:"refs(first: 100, refPrefix: \"refs/heads/\")"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

// Branches returns a list of branches for the given repository.
func (c *Client) Branches(owner string, name string) ([]vcs.Branch, error) {
	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
	}

	if err := c.queryWithRetry(context.Background(), &branchesQuery, variables); err != nil {
		return nil, err
	}

	var branches []vcs.Branch //nolint
	for _, node := range branchesQuery.Repository.Refs.Nodes {
		branches = append(branches, vcs.Branch{
			Name:       string(node.Name),
			LastCommit: commitFromQL(node.Target.Commit),
		})
	}

	return branches, nil
}
