package vcs

import (
	"time"
)

// PullRequest represents a pull request.
type PullRequest struct {
	ID        int
	Body      string
	Title     string
	Labels    Labels
	CreatedAt time.Time
}
