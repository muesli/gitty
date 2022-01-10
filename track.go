package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/muesli/gitty/vcs"
)

const maxTrackStatCount = 99

type trackStat struct {
	Outdated bool
	Ahead    int
	Behind   int
}

func (s *trackStat) Render() string {
	genericStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGray))
	outdatedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorRed))
	statCountStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorGreen)).Width(4).Align(lipgloss.Right)
	statCountWarnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorYellow)).Width(4).Align(lipgloss.Right)
	if s == nil {
		return genericStyle.Render(" ") + statCountStyle.Render(" ") + statCountStyle.Render(" ")
	} else {
		str := ""
		if s.Outdated {
			str += outdatedStyle.Render("↻")
		} else {
			str += genericStyle.Render(" ")
		}
		if s.Ahead > 0 || s.Behind > 0 {
			str += statCountWarnStyle.Render(s.AheadString())
			str += statCountWarnStyle.Render(s.BehindString())
		} else {
			str += statCountStyle.Render(s.AheadString())
			str += statCountStyle.Render(s.BehindString())
		}
		return str
	}
}

func (s trackStat) AheadString() string {
	if s.Ahead == 0 {
		return "↑"
	} else if s.Ahead > maxTrackStatCount {
		return fmt.Sprintf("%d+↑", maxTrackStatCount)
	} else {
		return fmt.Sprintf("%d↑", s.Ahead)
	}
}

func (s trackStat) BehindString() string {
	if s.Behind == 0 {
		return "↓"
	} else if s.Behind > maxTrackStatCount {
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
		} else {
			if result, err = getTrackStat(repo, b, &remoteBranch); err != nil {
				result = nil
			}
		}
		results[remoteBranch.Name] = result
	}
	return results, nil
}

func getTrackStat(repo *git.Repository, rawLocalBranch *config.Branch, remoteBranch *vcs.Branch) (*trackStat, error) {
	if localBranch, err := repo.Branch(rawLocalBranch.Name); err != nil {
		return nil, err
	} else if localRef, err := repo.Reference(
		plumbing.NewBranchReferenceName(localBranch.Name), true,
	); err != nil {
		return nil, err
	} else if remoteRef, err := repo.Reference(
		plumbing.NewRemoteReferenceName(localBranch.Remote, remoteBranch.Name), true,
	); err != nil {
		return nil, err
	} else {
		stat := &trackStat{
			Outdated: false,
			Ahead:    0,
			Behind:   0,
		}

		if stat.Ahead, stat.Behind, err = calculateTrackCount(
			repo, localRef.Hash(), remoteRef.Hash(),
		); err != nil {
			return nil, err
		}

		if remoteRef.Hash().String() != remoteBranch.LastCommit.ID {
			// mark outdated, need `git fetch`
			stat.Outdated = true
		}
		return stat, nil
	}
}

func calculateTrackCount(repo *git.Repository, ref, base plumbing.Hash) (ahead, behind int, err error) {
	if ref == base {
		return 0, 0, nil
	}

	left, err := repo.CommitObject(ref)
	if err != nil {
		return 0, 0, err
	}
	right, err := repo.CommitObject(base)
	if err != nil {
		return 0, 0, err
	}

	commitMap := make(map[plumbing.Hash]bool)

	if err := iterateCommits(left, func(c *object.Commit) {
		commitMap[c.Hash] = true
	}); err != nil {
		return 0, 0, err
	}

	if err := iterateCommits(right, func(c *object.Commit) {
		if _, ok := commitMap[c.Hash]; !ok {
			behind++
		} else {
			commitMap[c.Hash] = false
		}
	}); err != nil {
		return 0, 0, err
	}

	for _, v := range commitMap {
		if v {
			ahead++
		}
	}
	return
}

func iterateCommits(c *object.Commit, fn func(c *object.Commit)) error {
	iter := object.NewCommitPreorderIter(c, map[plumbing.Hash]bool{}, []plumbing.Hash{})
	defer iter.Close()
	for {
		if curr, err := iter.Next(); err == io.EOF {
			break
		} else if err != nil {
			return err
		} else {
			fn(curr)
		}
	}
	return nil
}
