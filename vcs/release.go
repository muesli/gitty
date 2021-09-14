package vcs

import (
	"time"
)

type Release struct {
	Name         string
	TagName      string
	PublishedAt  time.Time
	URL          string
	CommitsSince []Commit
}
