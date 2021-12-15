package vcs

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
