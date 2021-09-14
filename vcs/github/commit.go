package github

import (
	"context"
	"time"

	"github.com/muesli/gitty/vcs"
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

func (c *Client) History(owner string, name string, max int, since time.Time) ([]vcs.Commit, error) {
	var commits []vcs.Commit

	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
		"since": githubv4.GitTimestamp{Time: since},
	}

	// if err := client.Query(context.Background(), &historyQuery, variables); err != nil {
	if err := c.queryWithRetry(context.Background(), &historyQuery, variables); err != nil {
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

func CommitFromQL(commit QLCommit) vcs.Commit {
	return vcs.Commit{
		ID:              string(commit.OID),
		MessageHeadline: string(commit.MessageHeadline),
		CommittedAt:     commit.CommittedDate.Time,
		Author:          string(commit.Author.User.Login),
	}
}
