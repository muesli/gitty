package github

import (
	"context"

	"github.com/muesli/gitty/vcs"
	"github.com/shurcooL/githubv4"
)

type QLUser struct {
	Login     githubv4.String
	Name      githubv4.String
	AvatarURL githubv4.String
	URL       githubv4.String
}

var viewerQuery struct {
	Viewer struct {
		Login githubv4.String
	}
}

func (c *Client) GetUsername() (string, error) {
	if err := c.queryWithRetry(context.Background(), &viewerQuery, nil); err != nil {
		return "", err
	}

	return string(viewerQuery.Viewer.Login), nil
}

func UserFromQL(user QLUser) vcs.User {
	return vcs.User{
		Login:     string(user.Login),
		Name:      string(user.Name),
		AvatarURL: string(user.AvatarURL),
		URL:       string(user.URL),
	}
}
