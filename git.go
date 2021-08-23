package main

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

func repoOrigin(path string) (string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return "", err
	}

	remotes, err := r.Remotes()
	if err != nil {
		return "", err
	}

	for _, v := range remotes {
		if v.Config().Name == "origin" {
			return v.Config().URLs[0], nil
		}
	}

	return "", fmt.Errorf("no remote found")
}
