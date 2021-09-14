package vcs

import (
	"time"
)

type PullRequest struct {
	ID        int
	Body      string
	Title     string
	Labels    Labels
	CreatedAt time.Time
}
