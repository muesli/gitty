package vcs

import (
	"time"
)

// Issue represents an issue.
type Issue struct {
	ID        int
	Body      string
	Title     string
	Labels    Labels
	CreatedAt time.Time
}
