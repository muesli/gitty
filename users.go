package main

import (
	"context"

	"github.com/shurcooL/githubv4"
)

type QLUser struct {
	Login     githubv4.String
	Name      githubv4.String
	AvatarURL githubv4.String
	URL       githubv4.String
}

type User struct {
	Login     string
	Name      string
	AvatarURL string
	URL       string
}

var viewerQuery struct {
	Viewer struct {
		Login githubv4.String
	}
}

func getUsername() (string, error) {
	if err := queryWithRetry(context.Background(), &viewerQuery, nil); err != nil {
		return "", err
	}

	return string(viewerQuery.Viewer.Login), nil
}

func UserFromQL(user QLUser) User {
	return User{
		Login:     string(user.Login),
		Name:      string(user.Name),
		AvatarURL: string(user.AvatarURL),
		URL:       string(user.URL),
	}
}
