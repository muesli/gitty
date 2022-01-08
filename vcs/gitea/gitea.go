package gitea

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/muesli/gitty/vcs"
)

// Client is a gitea client.
type Client struct {
	api  *gitea.Client
	host string
}

// NewClient returns a new gitea client.
func NewClient(baseURL, token string, preverified bool) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("can't parse URL: %v", err)
	}
	u.Scheme = "https"

	client, err := gitea.NewClient(u.String(), gitea.SetToken(token))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	if !preverified {
		_, _, err := client.ServerVersion()
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		api:  client,
		host: baseURL,
	}, nil
}

// GetUsername returns the username of the authenticated user.
func (c *Client) GetUsername() (string, error) {
	u, _, err := c.api.GetMyUserInfo()
	if err != nil {
		return "", err
	}

	return u.UserName, nil
}

// Issues returns a list of issues for the given repository.
func (c *Client) Issues(owner string, name string) ([]vcs.Issue, error) {
	var i []vcs.Issue

	page := 1
	for {
		issues, _, err := c.api.ListRepoIssues(owner, name, gitea.ListIssueOption{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: 250,
			},
			State: gitea.StateOpen,
		})
		if err != nil {
			return nil, err
		}

		for _, v := range issues {
			issue := vcs.Issue{
				ID:        int(v.ID),
				Title:     v.Title,
				CreatedAt: v.Created,
			}
			for _, l := range v.Labels {
				issue.Labels = append(issue.Labels, vcs.Label{
					Name:  l.Name,
					Color: "#" + l.Color,
				})
			}
			i = append(i, issue)
		}

		page++
		if len(issues) == 0 {
			break
		}
	}

	return i, nil
}

// PullRequests returns a list of pull requests for the given repository.
func (c *Client) PullRequests(owner string, name string) ([]vcs.PullRequest, error) {
	var i []vcs.PullRequest

	page := 1
	for {
		prs, _, err := c.api.ListRepoPullRequests(owner, name, gitea.ListPullRequestsOptions{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: 250,
			},
			State: gitea.StateOpen,
		})
		if err != nil {
			return nil, err
		}

		for _, v := range prs {
			pr := vcs.PullRequest{
				ID:        int(v.ID),
				Title:     v.Title,
				CreatedAt: *v.Created,
			}
			for _, l := range v.Labels {
				pr.Labels = append(pr.Labels, vcs.Label{
					Name:  l.Name,
					Color: "#" + l.Color,
				})
			}
			i = append(i, pr)
		}

		page++
		if len(prs) == 0 {
			break
		}
	}

	return i, nil
}

// Repository returns the repository with the given name.
func (c *Client) Repository(owner string, name string) (vcs.Repo, error) {
	p, _, err := c.api.GetRepo(owner, name)
	if err != nil {
		return vcs.Repo{}, err
	}

	r := c.repoFromAPI(p)
	return r, nil
}

// Repositories returns a list of repositories for the given user.
func (c *Client) Repositories(owner string) ([]vcs.Repo, error) {
	var repos []vcs.Repo

	page := 1
	for {
		p, _, err := c.api.ListOrgRepos(owner, gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: 250,
			},
		})
		if err != nil {
			break
		}

		for _, v := range p {
			repos = append(repos, c.repoFromAPI(v))
		}

		page++
		if len(p) == 0 {
			break
		}
	}

	page = 0
	for {
		p, _, err := c.api.ListUserRepos(owner, gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: 250,
			},
		})
		if err != nil {
			break
		}

		for _, v := range p {
			repos = append(repos, c.repoFromAPI(v))
		}

		page++
		if len(p) == 0 {
			break
		}
	}

	return repos, nil //nolint
}

// Branches returns a list of branches for the given repository.
func (c *Client) Branches(owner string, name string) ([]vcs.Branch, error) {
	var i []vcs.Branch
	opts := gitea.ListRepoBranchesOptions{
		ListOptions: gitea.ListOptions{
			PageSize: 250,
		},
	}
	for {
		opts.Page++
		branches, _, err := c.api.ListRepoBranches(owner, name, opts)
		if err != nil {
			return nil, err
		}
		if len(branches) == 0 {
			break
		}

		for _, v := range branches {
			branch := vcs.Branch{
				Name: v.Name,
				LastCommit: vcs.Commit{
					ID:              v.Commit.ID,
					MessageHeadline: trimMessage(v.Commit.Message),
					CommittedAt:     v.Commit.Timestamp,
					Author:          v.Commit.Author.UserName,
				},
			}
			i = append(i, branch)
		}
	}

	return i, nil
}

// History returns a list of commits for the given repository.
func (c *Client) History(repo vcs.Repo, max int, since time.Time) ([]vcs.Commit, error) {
	var commits []vcs.Commit

	page := 1
	for {
		opt := gitea.ListCommitOptions{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: 250,
			},
		}
		h, _, err := c.api.ListRepoCommits(repo.Owner, repo.Name, opt)
		if err != nil {
			return nil, err
		}

		var brk bool
		for _, v := range h {
			if v.Created.Before(since) {
				brk = true
				break
			}
			commits = append(commits, vcs.Commit{
				ID:              v.SHA,
				MessageHeadline: trimMessage(v.RepoCommit.Message),
				CommittedAt:     v.Created,
				Author:          v.Author.UserName,
			})
		}
		if brk {
			break
		}

		page++
		if len(h) == 0 {
			break
		}
		if max > 0 && len(commits) >= max {
			break
		}
	}

	return commits, nil
}

// IssueURL returns the URL to the issue with the given number.
func (c *Client) IssueURL(owner string, name string, number int) string {
	i, _, err := c.api.GetIssue(owner, name, int64(number))
	if err == nil {
		return i.HTMLURL
	}

	p, _, err := c.api.GetPullRequest(owner, name, int64(number))
	if err == nil {
		return p.HTMLURL
	}

	return ""
}

func (c *Client) repoFromAPI(p *gitea.Repository) vcs.Repo {
	var release vcs.Release
	r, _, err := c.api.ListReleases(p.Owner.UserName, p.Name, gitea.ListReleasesOptions{})
	if err == nil && len(r) > 0 {
		release = vcs.Release{
			Name:        r[0].Title,
			TagName:     r[0].TagName,
			PublishedAt: r[0].CreatedAt,
		}
	}

	return vcs.Repo{
		Owner:         p.Owner.UserName,
		Name:          p.Name,
		NameWithOwner: p.FullName,
		URL:           p.HTMLURL,
		Description:   p.Description,
		Stargazers:    p.Stars,
		Watchers:      p.Watchers,
		Forks:         p.Forks,
		// Commits:       p.Statistics.CommitCount,
		LastRelease: release,
	}
}

func trimMessage(s string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "\n") {
		return strings.Split(s, "\n")[0]
	}

	return s
}
