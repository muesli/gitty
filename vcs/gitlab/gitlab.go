package gitlab

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/muesli/gamut"
	"github.com/muesli/gitty/vcs"
	"github.com/xanzy/go-gitlab"
)

// Client is a client for GitLab.
type Client struct {
	api         *gitlab.Client
	host        string
	colors      map[string]int
	labelColors map[string]string
}

// NewClient returns a new GitLab client.
func NewClient(baseURL, token string, preverified bool) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("can't parse URL: %v", err)
	}
	u.Path = path.Join(u.Path, "/api/v4")
	u.Scheme = "https"

	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(u.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	if !preverified {
		_, _, err = client.Version.GetVersion()
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		api:         client,
		host:        baseURL,
		colors:      map[string]int{},
		labelColors: map[string]string{},
	}, nil
}

// GetUsername returns the username of the authenticated user.
func (c *Client) GetUsername() (string, error) {
	u, _, err := c.api.Users.CurrentUser()
	if err != nil {
		return "", err
	}

	return u.Username, nil
}

// Issues returns a list of issues for the given repository.
func (c *Client) Issues(owner string, name string) ([]vcs.Issue, error) {
	var i []vcs.Issue

	page := 1
	for {
		issues, resp, err := c.api.Issues.ListProjectIssues(owner+"/"+name,
			&gitlab.ListProjectIssuesOptions{
				ListOptions: gitlab.ListOptions{
					Page:    page,
					PerPage: 250,
				},
				State: gitlab.String("opened"),
			})
		if err != nil {
			return nil, err
		}

		for _, v := range issues {
			issue := vcs.Issue{
				ID:        v.IID,
				Title:     v.Title,
				CreatedAt: *v.CreatedAt,
			}
			for _, l := range v.Labels {
				issue.Labels = append(issue.Labels, vcs.Label{
					Name:  l,
					Color: c.colorForLabel(l),
				})
			}
			i = append(i, issue)
		}

		page++
		if resp.TotalPages == page || len(issues) == 0 {
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
		prs, resp, err := c.api.MergeRequests.ListProjectMergeRequests(owner+"/"+name,
			&gitlab.ListProjectMergeRequestsOptions{
				ListOptions: gitlab.ListOptions{
					Page:    page,
					PerPage: 250,
				},
				State: gitlab.String("opened"),
			})
		if err != nil {
			return nil, err
		}

		for _, v := range prs {
			pr := vcs.PullRequest{
				ID:        v.IID,
				Title:     v.Title,
				CreatedAt: *v.CreatedAt,
			}
			for _, l := range v.Labels {
				pr.Labels = append(pr.Labels, vcs.Label{
					Name:  l,
					Color: c.colorForLabel(l),
				})
			}
			i = append(i, pr)
		}

		page++
		if resp.TotalPages == page || len(prs) == 0 {
			break
		}
	}

	return i, nil
}

// Repository returns the repository with the given name.
func (c *Client) Repository(owner string, name string) (vcs.Repo, error) {
	p, _, err := c.api.Projects.GetProject(owner+"/"+name, nil)
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
		p, resp, err := c.api.Groups.ListGroupProjects(owner, &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: 250,
			},
		})
		if err != nil {
			break
		}

		for _, v := range p {
			repos = append(repos, c.repoFromAPI(v))
		}

		page++
		if resp.TotalPages == page || len(p) == 0 {
			break
		}
	}

	page = 0
	for {
		p, resp, err := c.api.Projects.ListUserProjects(owner, &gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: 250,
			},
		})
		if err != nil {
			break
		}

		for _, v := range p {
			repos = append(repos, c.repoFromAPI(v))
		}

		page++
		if resp.TotalPages == page || len(p) == 0 {
			break
		}
	}

	return repos, nil //nolint
}

// Branches returns a list of branches for the given repository.
func (c *Client) Branches(owner string, name string) ([]vcs.Branch, error) {
	var i []vcs.Branch
	opts := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 250,
		},
	}
	for {
		opts.Page++
		branches, _, err := c.api.Branches.ListBranches(owner+"/"+name, opts)
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
					MessageHeadline: v.Commit.Title,
					CommittedAt:     *v.Commit.CommittedDate,
					Author:          v.Commit.CommitterName,
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
		opt := gitlab.ListCommitsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: 250,
			},
		}
		if !since.IsZero() {
			opt.Since = &since
		}
		h, resp, err := c.api.Commits.ListCommits(repo.NameWithOwner, &opt)
		if err != nil {
			return nil, err
		}

		for _, v := range h {
			commits = append(commits, vcs.Commit{
				ID:              v.ID,
				MessageHeadline: strings.ReplaceAll(v.Title, "\u00A0", " "),
				CommittedAt:     *v.CommittedDate,
				Author:          v.AuthorName,
			})
		}

		page++
		if resp.TotalPages == page || len(h) == 0 {
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
	i, _, err := c.api.Issues.GetIssue(owner+"/"+name, number)
	if err == nil {
		return i.WebURL
	}

	p, _, err := c.api.MergeRequests.GetMergeRequest(owner+"/"+name, number, nil)
	if err == nil {
		return p.WebURL
	}

	return ""
}

func (c *Client) repoFromAPI(p *gitlab.Project) vcs.Repo {
	var release vcs.Release
	r, _, err := c.api.Releases.ListReleases(p.PathWithNamespace, nil)
	if err == nil && len(r) > 0 {
		release = vcs.Release{
			Name:        r[0].Name,
			TagName:     r[0].TagName,
			PublishedAt: *r[0].CreatedAt,
		}
	}

	return vcs.Repo{
		Owner:         p.Namespace.Path,
		Name:          p.Name,
		NameWithOwner: p.PathWithNamespace,
		URL:           p.WebURL,
		Description:   p.Description,
		Stargazers:    p.StarCount,
		Watchers:      0,
		Forks:         p.ForksCount,
		// Commits:       p.Statistics.CommitCount,
		LastRelease: release,
	}
}

func (c *Client) colorForLabel(label string) string {
	color, ok := c.labelColors[label]
	if ok {
		return color
	}

	if len(c.colors) == 0 {
		cc, err := gamut.Generate(12, gamut.PastelGenerator{})
		if err != nil {
			return "#333333"
		}

		for _, v := range cc {
			c.colors[gamut.ToHex(v)] = 0
		}
	}

	least := -1
	for k, v := range c.colors {
		if v < least || least < 0 {
			least = v
			color = k
		}
	}

	c.colors[color]++
	c.labelColors[label] = color

	return c.labelColors[label]
}
