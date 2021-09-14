package vcs

import (
	"time"
)

type Issue struct {
	ID        int
	Body      string
	Title     string
	Labels    Labels
	CreatedAt time.Time
}
