package vcs

import (
	"time"
)

type Commit struct {
	ID              string
	MessageHeadline string
	CommittedAt     time.Time
	Author          string
}
