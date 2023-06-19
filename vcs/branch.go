package vcs

// Branch represents a git branch.
type Branch struct {
	Name       string
	LastCommit Commit
	URL        string
}
