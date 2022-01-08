package vcs

import (
	"time"
)

// Release represents a release.
type Release struct {
	Name         string
	TagName      string
	PublishedAt  time.Time
	URL          string
	CommitsSince []Commit
}
