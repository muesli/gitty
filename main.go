package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/skratchdot/open-golang/open"
)

var (
	Version   = ""
	CommitSHA = ""

	maxBranches     = flag.Int("max-branches", 10, "Max amount of active branches to show")
	maxCommits      = flag.Int("max-commits", 10, "Max amount of commits to show")
	maxIssues       = flag.Int("max-issues", 10, "Max amount of issues to show")
	maxPullRequests = flag.Int("max-pull-requests", 10, "Max amount of pull requests to show")
	maxBranchAge    = flag.Int("max-branch-age", 28, "Max age of a branch in days to be considered active")
	minNewCommits   = flag.Int("min-new-commits", 1, "Min amount of new commits for a repo to be considered new")
	skipStaleRepos  = flag.Bool("skip-stale-repos", true, "Skip repos without new activity")
	withCommits     = flag.Bool("with-commits", false, "Show new commits")
	allProjects     = flag.Bool("all-projects", false, "Retrieve information for all source repositories")

	version = flag.Bool("version", false, "display version")

	theme Theme
)

func parseRepository() {
	arg := "."
	num := 0

	// parse args
	args := flag.Args()
	if len(args) > 0 {
		arg = args[0]
		args = args[1:]

		if len(args) == 0 {
			// only one arg provided. Is it an issue/pr number?
			var err error
			num, err = strconv.Atoi(arg)
			if err == nil {
				arg = "."
			}
		}
	}
	if len(args) > 0 {
		var err error
		num, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		args = args[1:]
	}

	// parse GitHub URL from args
	var u string
	origin, err := repoOrigin(arg)
	if err != nil {
		if !strings.Contains(arg, "github.com/") {
			fmt.Println(err)
			os.Exit(1)
		}

		u = cleanupUrl(arg)
		origin = u + ".git"
	} else {
		u, err = githubURL(arg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	p := strings.Split(u, "/")
	owner, name := p[3], p[4]

	// launched with issue/pr number?
	if num > 0 {
		iu := fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, name, num)
		if err := open.Start(iu); err != nil {
			fmt.Println("URL:", iu)
		}
		os.Exit(0)
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorCyan))
	tooltipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorTooltip))

	_ = origin
	// fmt.Println(tooltipStyle.Render("🏠 Remote ") + headerStyle.Render(origin))
	// fmt.Println(tooltipStyle.Render("🔖 Website ") + headerStyle.Render(u))
	fmt.Println(tooltipStyle.Render("🏠 Repository ") + headerStyle.Render(u))

	// fetch issues
	is := make(chan []Issue)
	go func() {
		i, err := issues(owner, name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		is <- i
	}()

	// fetch pull requests
	prs := make(chan []PullRequest)
	go func() {
		p, err := pullRequests(owner, name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		prs <- p
	}()

	// fetch active branches
	brs := make(chan []Branch)
	go func() {
		b, err := branches(owner, name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		brs <- b
	}()

	// fetch commit history
	repo := make(chan Repo)
	go func() {
		r, err := repository(owner, name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		r.LastRelease.CommitsSince, err = history(r.Owner, r.Name, r.LastRelease.PublishedAt)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		repo <- r
	}()

	printIssues(<-is)
	printPullRequests(<-prs)
	printBranches(<-brs)
	printCommits(<-repo)
}

func parseAllProjects() {
	repos, err := repositories(username)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}
	mut := &sync.Mutex{}
	var rr []Repo

	// repos with a release
	for _, repo := range reposWithRelease(repos) {
		wg.Add(1)

		go func(repo Repo) {
			var err error
			repo.LastRelease.CommitsSince, err = history(repo.Owner, repo.Name, repo.LastRelease.PublishedAt)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			mut.Lock()
			rr = append(rr, repo)
			mut.Unlock()

			wg.Done()
		}(repo)
	}

	wg.Wait()
	fmt.Printf("%d repositories with a release:\n", len(rr))

	sort.Slice(rr, func(i, j int) bool {
		if rr[i].LastRelease.PublishedAt.Equal(rr[j].LastRelease.PublishedAt) {
			return strings.Compare(rr[i].Name, rr[j].Name) < 0
		}
		return rr[i].LastRelease.PublishedAt.After(rr[j].LastRelease.PublishedAt)
	})

	for _, repo := range rr {
		repoRelease(repo)
	}
}

func main() {
	flag.Parse()
	if *version {
		if len(CommitSHA) > 7 {
			CommitSHA = CommitSHA[:7]
		}
		if Version == "" {
			Version = "(built from source)"
		}

		fmt.Printf("gitty %s", Version)
		if len(CommitSHA) > 0 {
			fmt.Printf(" (%s)", CommitSHA)
		}

		fmt.Println()
		os.Exit(0)
	}

	initTheme()

	if err := initGitHub(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *allProjects {
		parseAllProjects()
		os.Exit(0)
	}

	parseRepository()
}

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
