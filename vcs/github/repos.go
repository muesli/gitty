package github

import (
	"context"

	"github.com/muesli/gitty/vcs"
	"github.com/shurcooL/githubv4"
)

var reposQuery struct {
	User struct {
		Login        githubv4.String
		Repositories struct {
			TotalCount githubv4.Int
			Edges      []struct {
				Cursor githubv4.String
				Node   struct {
					QLRepository
				}
			}
		} `graphql:"repositories(first: 100, after:$after isFork: false, ownerAffiliations: OWNER, orderBy: {field: CREATED_AT, direction: DESC})"`
	} `graphql:"repositoryOwner(login:$username)"`
}

var repoQuery struct {
	Repository QLRepository `graphql:"repository(owner: $owner, name: $name)"`
}

type QLRepository struct {
	Owner struct {
		Login githubv4.String
	}
	Name           githubv4.String
	NameWithOwner  githubv4.String
	URL            githubv4.String
	Description    githubv4.String
	IsPrivate      githubv4.Boolean
	ForkCount      githubv4.Int
	StargazerCount githubv4.Int

	Watchers struct {
		TotalCount githubv4.Int
	}

	BranchEntity struct {
		Commits struct {
			History struct {
				TotalCount githubv4.Int
			}
		} `graphql:"... on Commit"`
	} `graphql:"object(expression: \"HEAD\")"`

	Releases QLRelease `graphql:"releases(first: 10, orderBy: {field: CREATED_AT, direction: DESC})"`
}

func (c *Client) Repository(owner string, name string) (vcs.Repo, error) {
	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
	}

	if err := c.queryWithRetry(context.Background(), &repoQuery, variables); err != nil {
		return vcs.Repo{}, err
	}

	repo := RepoFromQL(repoQuery.Repository)
	if len(repoQuery.Repository.Releases.Nodes) > 0 {
		repo.LastRelease = ReleaseFromQL(repoQuery.Repository.Releases)
	}

	return repo, nil
}

func (c *Client) Repositories(owner string) ([]vcs.Repo, error) {
	var after *githubv4.String
	var repos []vcs.Repo

	for {
		variables := map[string]interface{}{
			"username": githubv4.String(owner),
			"after":    after,
		}

		if err := c.queryWithRetry(context.Background(), &reposQuery, variables); err != nil {
			return nil, err
		}
		if len(reposQuery.User.Repositories.Edges) == 0 {
			break
		}

		for _, v := range reposQuery.User.Repositories.Edges {
			repo := RepoFromQL(v.Node.QLRepository)
			if len(v.Node.Releases.Nodes) > 0 {
				repo.LastRelease = ReleaseFromQL(v.Node.Releases)
			}

			repos = append(repos, repo)

			after = &v.Cursor
		}
	}

	return repos, nil
}

func RepoFromQL(repo QLRepository) vcs.Repo {
	return vcs.Repo{
		Owner:         string(repo.Owner.Login),
		Name:          string(repo.Name),
		NameWithOwner: string(repo.NameWithOwner),
		URL:           string(repo.URL),
		Description:   string(repo.Description),
		Stargazers:    int(repo.StargazerCount),
		Watchers:      int(repo.Watchers.TotalCount),
		Forks:         int(repo.ForkCount),
		Commits:       int(repo.BranchEntity.Commits.History.TotalCount),
	}
}
