package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var (
	username string
	client   *githubv4.Client
	// clientv3 *github.Client
)

func githubURL(path string) (string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return "", err
	}

	remotes, err := r.Remotes()
	if err != nil {
		return "", err
	}

	var u string
	var rn string
	for _, v := range remotes {
		if !strings.Contains(v.Config().URLs[0], "github.com") {
			continue
		}

		if (v.Config().Name == "origin" && rn != "origin") ||
			rn == "" {
			rn = v.Config().Name
			u = v.Config().URLs[0]
		}
	}

	if u == "" {
		return "", fmt.Errorf("no GitHub remote found")
	}

	return cleanupUrl(u), nil
}

func cleanupUrl(u string) string {
	u = strings.TrimSuffix(u, ".git")
	p := strings.Split(u, "github.com")
	p[1] = strings.TrimPrefix(p[1], ":")
	p[1] = strings.TrimPrefix(p[1], "/")

	return fmt.Sprintf("https://github.com/%s", p[1])
}

func initGitHub() error {
	var httpClient *http.Client
	token := os.Getenv("GITHUB_TOKEN")
	if len(token) == 0 {
		token = os.Getenv("GITTY_TOKEN")
		if len(token) == 0 {
			return fmt.Errorf("Please set a GITHUB_TOKEN or GITTY_TOKEN env var")
		}
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	httpClient = oauth2.NewClient(context.Background(), ts)
	client = githubv4.NewClient(httpClient)

	/*
		tc := oauth2.NewClient(context.Background(), ts)
		clientv3 = github.NewClient(tc)
	*/

	var err error
	username, err = getUsername()
	if err != nil {
		return fmt.Errorf("Can't retrieve GitHub profile: %s", err)
	}

	return nil
}

func queryWithRetry(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	if err := client.Query(context.Background(), q, variables); err != nil {
		if strings.Contains(err.Error(), "abuse-rate-limits") {
			time.Sleep(time.Minute)
			return queryWithRetry(ctx, q, variables)
		}

		return err
	}

	return nil
}
