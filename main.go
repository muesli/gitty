package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/gitty/vcs"
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
	namespace       = flag.String("namespace", "", "User/organization name when using --all-projects")

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

	// parse URL from args
	host, owner, name, err := parseRepo(arg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Printf("Host: %s, Owner: %s, Name: %s\n", host, owner, name)

	// guess appropriate API client from hostname
	client, err := guessClient(host)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// launched with issue/pr number?
	if num > 0 {
		iu := client.IssueURL(owner, name, num)
		if len(iu) == 0 {
			fmt.Printf("Issue/PR %d not found\n", num)
			os.Exit(1)
		}
		if err := open.Start(iu); err != nil {
			fmt.Println("URL:", iu)
		}
		os.Exit(0)
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorCyan))
	tooltipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.colorTooltip))

	// fmt.Println(tooltipStyle.Render("🏠 Remote ") + headerStyle.Render(origin))
	// fmt.Println(tooltipStyle.Render("🔖 Website ") + headerStyle.Render(u))
	fmt.Println(tooltipStyle.Render("🏠 Repository ") + headerStyle.Render("https://"+host+"/"+owner+"/"+name))

	// fetch issues
	is := make(chan []vcs.Issue)
	go func() {
		i, err := client.Issues(owner, name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		is <- i
	}()

	// fetch pull requests
	prs := make(chan []vcs.PullRequest)
	go func() {
		p, err := client.PullRequests(owner, name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		prs <- p
	}()

	// fetch active branches
	brs := make(chan []vcs.Branch)
	go func() {
		b, err := client.Branches(owner, name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		brs <- b
	}()

	// fetch commit history
	repo := make(chan vcs.Repo)
	go func() {
		r, err := client.Repository(owner, name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		r.LastRelease.CommitsSince, err = client.History(r.Owner, r.Name, *maxCommits, r.LastRelease.PublishedAt)
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
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Please provide the hostname of a git provider, e.g. github.com")
		os.Exit(1)
	}

	client, err := guessClient(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *namespace == "" {
		u, err := client.GetUsername()
		if err != nil {
			fmt.Printf("Can't retrieve profile: %s\n", err)
			os.Exit(1)
		}
		*namespace = u
	}

	repos, err := client.Repositories(*namespace)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}
	mut := &sync.Mutex{}
	var rr []vcs.Repo

	// repos with a release
	for _, repo := range vcs.ReposWithRelease(repos) {
		wg.Add(1)

		go func(repo vcs.Repo) {
			var err error
			repo.LastRelease.CommitsSince, err = client.History(repo.Owner, repo.Name, *maxCommits, repo.LastRelease.PublishedAt)
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

func printVersion() {
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
}

func main() {
	flag.Parse()
	if *version {
		printVersion()
		os.Exit(0)
	}

	initTheme()

	if *allProjects {
		parseAllProjects()
		os.Exit(0)
	}

	parseRepository()
}
