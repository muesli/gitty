package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// Client is a GitHub client.
type Client struct {
	api *githubv4.Client
}

// NewClient creates a new GitHub client.
func NewClient(token string) (*Client, error) {
	var httpClient *http.Client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	httpClient = oauth2.NewClient(context.Background(), ts)
	client := githubv4.NewClient(httpClient)

	c := &Client{
		api: client,
	}

	return c, nil
}

func (c *Client) queryWithRetry(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	if err := c.api.Query(ctx, q, variables); err != nil {
		if strings.Contains(err.Error(), "abuse-rate-limits") {
			time.Sleep(time.Minute)
			return c.queryWithRetry(ctx, q, variables)
		}

		return err
	}

	return nil
}

// IssueURL returns the URL to the issue with the given number.
func (c *Client) IssueURL(owner string, name string, number int) string {
	return fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, name, number)
}
