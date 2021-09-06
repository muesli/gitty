package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

func ago(t time.Time) string {
	s := humanize.Time(t)
	if strings.Contains(s, "minute") || strings.Contains(s, "second") {
		return "now"
	}

	s = strings.TrimSuffix(s, " ago")
	s = strings.ReplaceAll(s, "years", "y")
	s = strings.ReplaceAll(s, "year", "y")
	s = strings.ReplaceAll(s, "months", "m")
	s = strings.ReplaceAll(s, "month", "m")
	s = strings.ReplaceAll(s, "weeks", "w")
	s = strings.ReplaceAll(s, "week", "w")
	s = strings.ReplaceAll(s, "days", "d")
	s = strings.ReplaceAll(s, "day", "d")
	s = strings.ReplaceAll(s, "hours", "h")
	s = strings.ReplaceAll(s, "hour", "h")
	s = strings.ReplaceAll(s, " ", "")

	return s
}

func pluralize(count int, singular string, plural string) string {
	if count == 0 {
		return fmt.Sprintf("No %s", plural)
	} else if count == 1 {
		return fmt.Sprintf("1 %s", singular)
	} else {
		return fmt.Sprintf("%d %s", count, plural)
	}
}
