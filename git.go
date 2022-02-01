package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/kevinburke/ssh_config"
	"github.com/muesli/gitty/vcs"
	"github.com/muesli/gitty/vcs/gitea"
	"github.com/muesli/gitty/vcs/github"
	"github.com/muesli/gitty/vcs/gitlab"
)

const (
	originRemote   = "origin"
	upstreamRemote = "upstream"
)

// Client defines the set of methods required from a git provider.
type Client interface {
	Issues(owner string, name string) ([]vcs.Issue, error)
	PullRequests(owner string, name string) ([]vcs.PullRequest, error)
	Repository(owner string, name string) (vcs.Repo, error)
	Repositories(owner string) ([]vcs.Repo, error)
	Branches(owner string, name string) ([]vcs.Branch, error)
	History(repo vcs.Repo, max int, since time.Time) ([]vcs.Commit, error)

	GetUsername() (string, error)
	IssueURL(owner string, name string, number int) string
}

func tokenForHost(host string) string {
	token := os.Getenv("GITTY_TOKENS")

	tokens := strings.Split(token, ";")
	for _, t := range tokens {
		if !strings.Contains(t, "=") {
			continue
		}

		s := strings.Split(t, "=")
		k, v := s[0], s[1]
		if !strings.EqualFold(k, host) {
			continue
		}

		return strings.TrimSpace(v)
	}

	// fallback for old tokens
	if host == "github.com" {
		token = os.Getenv("GITTY_TOKEN")
		if len(token) > 0 {
			return token
		}
		token = os.Getenv("GITHUB_TOKEN")
		if len(token) > 0 {
			return token
		}
	}

	return ""
}

func guessClient(host string) (Client, error) {
	token := tokenForHost(host)
	if len(token) == 0 {
		return nil, fmt.Errorf("please set a GITTY_TOKENS env var for host " + host)
	}

	if strings.EqualFold(host, "github.com") {
		return github.NewClient(token)
	}
	if strings.EqualFold(host, "gitlab.com") {
		return gitlab.NewClient(host, token, true)
	}
	if strings.EqualFold(host, "gitea.com") {
		return gitea.NewClient(host, token, true)
	}
	if strings.EqualFold(host, "codeberg.org") {
		return gitea.NewClient(host, token, true)
	}
	if strings.Contains(host, "invent.kde.org") {
		return gitlab.NewClient(host, token, true)
	}

	var client Client
	var err error
	client, err = gitlab.NewClient(host, token, false)
	if err == nil {
		return client, nil
	}
	client, err = gitea.NewClient(host, token, false)
	if err == nil {
		return client, nil
	}
	// fmt.Println(err)

	return nil, fmt.Errorf("not a recognized git provider")
}

// remoteURL returns remote name and URL
func remoteURL(path string) (string, string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return "", "", err
	}

	remotes, err := r.Remotes()
	if err != nil {
		return "", "", err
	}

	var u string
	var rn string
	for _, v := range remotes {
		if (v.Config().Name == upstreamRemote && rn != upstreamRemote) ||
			rn == "" {
			rn = v.Config().Name
			u = v.Config().URLs[0]
		}
		if v.Config().Name == originRemote && rn != originRemote && rn != upstreamRemote {
			rn = v.Config().Name
			u = v.Config().URLs[0]
		}
	}

	if u == "" {
		return "", "", fmt.Errorf("no remote found")
	}

	u, err = cleanupURL(u)
	return rn, u, err
}

func cleanupURL(arg string) (string, error) {
	var sshURL bool

	if strings.Contains(arg, "://") {
		u, err := url.Parse(arg)
		if err == nil {
			host, _, err := net.SplitHostPort(u.Host)
			if err == nil {
				// strip port
				u.Host = host
				arg = u.String()
			}
		}
		arg = strings.Split(arg, "://")[1]
	} else {
		sshURL = true
		arg = strings.ReplaceAll(arg, ":", "/")
	}

	arg = "https://" + arg
	u, err := url.Parse(arg)
	if err != nil {
		return "", err
	}

	// SSH URLs support stanzas we need to resolve.
	if sshURL {
		nh := ssh_config.Get(u.Hostname(), "Hostname")
		if nh != "" {
			u.Host = nh
		}
	}

	u.Path = strings.TrimSuffix(u.Path, ".git")
	u.User = nil
	u.Scheme = "https"

	// support SSH connections to GitHub via port 443
	if u.Host == "ssh.github.com" {
		u.Host = "github.com"
	}

	return u.String(), nil
}

// parseRepo returns host, owner, repository name and remote name from a given path or URL.
func parseRepo(arg string) (string, string, string, string, error) {
	rn, u, err := remoteURL(arg)
	if err != nil {
		// not a local repo
		u, err = cleanupURL(arg)
		if err != nil {
			return "", "", "", "", err
		}
	}

	p := strings.Split(u, "/")
	if len(p) < 5 {
		return "", "", "", "", fmt.Errorf("does not look like a valid path or URL")
	}

	host, owner, name := p[2], p[3], strings.Join(p[4:], "/")
	return host, owner, name, rn, nil
}
