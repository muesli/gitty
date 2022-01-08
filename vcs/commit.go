package vcs

import (
	"time"
)

// Commit represents a git commit.
type Commit struct {
	ID              string
	MessageHeadline string
	CommittedAt     time.Time
	Author          string
}
