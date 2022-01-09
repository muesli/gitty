package bitbucketcloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/ktrysmt/go-bitbucket"
	"github.com/mitchellh/mapstructure"
	"github.com/muesli/gitty/vcs"
)

type Client struct {
	api *bitbucket.Client
}

func NewClient(token string) (*Client, error) {
	split := strings.SplitN(token, ":", 2)
	if len(split) != 2 {
		return nil, fmt.Errorf("failed to get username and app password for bitbucket. Make sure to provide both username and password separated by a colon")
	}
	client := bitbucket.NewBasicAuth(split[0], split[1])

	return &Client{
		api: client,
	}, nil
}

func (c *Client) GetUsername() (string, error) {
	user, err := c.api.User.Profile()
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

func (c *Client) Issues(owner string, name string) ([]vcs.Issue, error) {
	type IssueResponse struct {
		Values []struct {
			ID        float64
			Title     string
			CreatedOn string `mapstructure:"created_on"`
			Content   struct {
				Raw string
			}
		}
	}
	var i []vcs.Issue
	issueResponse, err := c.api.Repositories.Issues.Gets(&bitbucket.IssuesOptions{Owner: owner, RepoSlug: name})
	if err != nil {
		// TODO: Find out if issue tracker is disabled, return zero issues instead of the error
		if strings.Contains(err.Error(), "404") {
			return i, nil
		}
		return i, err
	}

	var issueResponseTyped IssueResponse
	err = mapstructure.Decode(issueResponse, &issueResponseTyped)
	if err != nil {
		return i, err
	}

	for _, issue := range issueResponseTyped.Values {
		createdAt, err := time.Parse("2006-01-02T15:04:05.999999Z07:00", issue.CreatedOn)
		if err != nil {
			return i, err
		}
		i = append(i, vcs.Issue{
			ID:        int(issue.ID),
			Title:     issue.Title,
			Body:      issue.Content.Raw,
			CreatedAt: createdAt,
		})
	}
	// TODO: some issues are considered closed, like duplicates
	return i, nil
}

func (c *Client) PullRequests(owner string, name string) ([]vcs.PullRequest, error) {
	type PullRequestResponse struct {
		Values []struct {
			ID        float64
			Title     string
			CreatedOn string `mapstructure:"created_on"`
			Content   struct {
				Raw string
			}
		}
	}
	var prs []vcs.PullRequest
	prResponse, err := c.api.Repositories.PullRequests.Gets(&bitbucket.PullRequestsOptions{Owner: owner, RepoSlug: name})
	if err != nil {
		return prs, err
	}

	var prResponseTyped PullRequestResponse
	err = mapstructure.Decode(prResponse, &prResponseTyped)
	if err != nil {
		return prs, err
	}

	for _, pr := range prResponseTyped.Values {
		createdAt, err := time.Parse("2006-01-02T15:04:05.999999Z07:00", pr.CreatedOn)
		if err != nil {
			return prs, err
		}
		prs = append(prs, vcs.PullRequest{
			ID:        int(pr.ID),
			Title:     pr.Title,
			Body:      pr.Content.Raw,
			CreatedAt: createdAt,
		})
	}
	return prs, nil
}

func (c *Client) Repository(owner string, name string) (vcs.Repo, error) {
	repo, err := c.api.Repositories.Repository.Get(&bitbucket.RepositoryOptions{
		Owner:    owner,
		RepoSlug: name,
	})
	if err != nil {
		return vcs.Repo{}, err
	}
	html, _ := repo.Links["html"].(map[string]interface{})["href"].(string)

	return vcs.Repo{
		Owner:         repo.Owner["display_name"].(string),
		Name:          repo.Name,
		NameWithOwner: "foo",
		URL:           html,
		Description:   repo.Description,
	}, nil
}

func (c *Client) Repositories(owner string) ([]vcs.Repo, error) {
	return nil, nil
}

func (c *Client) Branches(owner string, name string) ([]vcs.Branch, error) {
	type BranchResponse struct {
		Branches []struct {
			Name   string
			Target struct {
				Hash   string
				Author struct {
					User struct {
						DisplayName string `mapstructure:"display_name"`
					}
				}
				Date    string
				Message string
			}
		}
	}
	var branches []vcs.Branch
	branchesResponse, err := c.api.Repositories.Repository.ListBranches(&bitbucket.RepositoryBranchOptions{Owner: owner, RepoSlug: name})
	if err != nil {
		return branches, err
	}

	var branchesResponseTyped BranchResponse
	err = mapstructure.Decode(branchesResponse, &branchesResponseTyped)
	if err != nil {
		return branches, err
	}

	for _, branch := range branchesResponseTyped.Branches {
		date, err := time.Parse("2006-01-02T15:04:05.999999Z07:00", branch.Target.Date)
		if err != nil {
			return branches, err
		}
		branches = append(branches, vcs.Branch{
			Name: branch.Name,
			LastCommit: vcs.Commit{
				ID:              branch.Target.Hash,
				CommittedAt:     date,
				MessageHeadline: strings.SplitN(branch.Target.Message, "\n", 2)[0],
				Author:          branch.Target.Author.User.DisplayName,
			},
		})
	}
	return branches, nil
}

func (c *Client) History(repo vcs.Repo, max int, since time.Time) ([]vcs.Commit, error) {
	var com []vcs.Commit
	type CommitResponse struct {
		Values []struct {
			Date    string
			Message string
			Hash    string

			Author struct {
				User struct {
					DisplayName string `mapstructure:"display_name"`
				}
			}
		}
	}

	commitResponse, err := c.api.Repositories.Commits.GetCommits(&bitbucket.CommitsOptions{
		Owner:    repo.Owner,
		RepoSlug: repo.Name,
	})
	var commitResponseTyped CommitResponse
	err = mapstructure.Decode(commitResponse, &commitResponseTyped)
	if err != nil {
		return com, err
	}

	for _, commit := range commitResponseTyped.Values {
		date, err := time.Parse("2006-01-02T15:04:05.999999Z07:00", commit.Date)
		if err != nil {
			return com, err
		}
		com = append(com, vcs.Commit{
			ID: commit.Hash,
			MessageHeadline: strings.SplitN(commit.Message, "\n", 2)[0],
			CommittedAt: date,
			Author: commit.Author.User.DisplayName,
		})
	}

	return com, nil
}

func (c *Client) IssueURL(owner string, name string, number int) string {
	return fmt.Sprintf("https://bitbucket.org/%s/%s/%d", owner, name, number)
}
