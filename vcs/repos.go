package vcs

// Repo represents a repository.
type Repo struct {
	Owner         string
	Name          string
	NameWithOwner string
	URL           string
	Description   string
	Stargazers    int
	Watchers      int
	Forks         int
	Commits       int
	LastRelease   Release
}

// ReposWithRelease returns all the repos that have a release.
func ReposWithRelease(repos []Repo) []Repo {
	var r []Repo //nolint

	for _, repo := range repos {
		if repo.LastRelease.PublishedAt.IsZero() {
			continue
		}

		r = append(r, repo)
	}

	return r
}
