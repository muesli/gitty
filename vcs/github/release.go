package github

import (
	"github.com/muesli/gitty/vcs"
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

func ReleaseFromQL(release QLRelease) vcs.Release {
	return vcs.Release{
		Name:        string(release.Nodes[0].Name),
		TagName:     string(release.Nodes[0].TagName),
		PublishedAt: release.Nodes[0].PublishedAt.Time,
		URL:         string(release.Nodes[0].URL),
	}
}
