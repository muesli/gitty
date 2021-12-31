package main

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/revlist"
	"github.com/muesli/gitty/vcs"
)

const maxTrackStatCount = 99

type trackStat struct {
	Ahead  int
	Behind int
}

func (s trackStat) AheadString() string {
	if s.Ahead > maxTrackStatCount {
		return fmt.Sprintf("%d+↑", maxTrackStatCount)
	} else {
		return fmt.Sprintf("%d↑", s.Ahead)
	}
}

func (s trackStat) BehindString() string {
	if s.Behind > maxTrackStatCount {
		return fmt.Sprintf("%d+↓", maxTrackStatCount)
	} else {
		return fmt.Sprintf("%d↓", s.Behind)
	}
}

func getBranchTrackStats(path string, remote string, trackedRemoteBranches []vcs.Branch) (map[string]*trackStat, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	iter, err := repo.Branches()
	if err != nil {
		return nil, err
	}

	trackedBranchMap := make(map[string]*config.Branch)

	if err := iter.ForEach(func(branchRef *plumbing.Reference) error {
		if b, err := repo.Branch(branchRef.Name().Short()); err == nil {
			if b.Remote == remote {
				trackedBranchMap[b.Merge.Short()] = b
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	results := make(map[string]*trackStat, len(trackedRemoteBranches))
	for _, remoteBranch := range trackedRemoteBranches {
		var result *trackStat = nil
		if b, ok := trackedBranchMap[remoteBranch.Name]; !ok {
		} else if localBranch, err := repo.Branch(b.Name); err != nil {
		} else if localRef, err := repo.Reference(plumbing.NewBranchReferenceName(localBranch.Name), true); err != nil {
		} else if remoteRef, err := repo.Reference(plumbing.NewRemoteReferenceName(remote, remoteBranch.Name), true); err != nil {
		} else if remoteRef.Hash().String() != remoteBranch.LastCommit.ID {
			// TODO: mark out of sync
		} else {
			result = getTrackStat(repo, [2]plumbing.Hash{remoteRef.Hash(), localRef.Hash()})
		}
		results[remoteBranch.Name] = result
	}
	return results, nil
}

func getAhead(repo *git.Repository, ref plumbing.Hash, base plumbing.Hash) (int, error) {
	hashes, err := revlist.Objects(repo.Storer,
		[]plumbing.Hash{ref},
		[]plumbing.Hash{base},
	)
	if err != nil {
		return 0, err
	}
	counter := 0
	for _, hash := range hashes {
		if c, err := repo.CommitObject(hash); err == nil && c != nil {
			counter++
		}
	}
	return counter, nil
}

func getTrackStat(repo *git.Repository, hashes [2]plumbing.Hash) *trackStat {
	if ahead, err := getAhead(repo, hashes[0], hashes[1]); err != nil {
		return nil
	} else if behind, err := getAhead(repo, hashes[1], hashes[0]); err != nil {
		return nil
	} else {
		return &trackStat{
			Ahead:  ahead,
			Behind: behind,
		}
	}
}
