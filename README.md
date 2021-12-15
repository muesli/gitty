# gitty

[![Latest Release](https://img.shields.io/github/release/muesli/gitty.svg)](https://github.com/muesli/gitty/releases)
[![Build Status](https://github.com/muesli/gitty/workflows/build/badge.svg)](https://github.com/muesli/gitty/actions)
[![Go ReportCard](https://goreportcard.com/badge/muesli/gitty)](https://goreportcard.com/report/muesli/gitty)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/muesli/gitty)

`gitty` is a smart little CLI helper for git projects, that shows you all the
relevant issues, pull requests and changes at a quick glance, right on the
command-line. It currently supports the GitHub & GitLab APIs.

![Screenshot](screenshot.png)

## Installation

### Packages

#### Linux
- Arch Linux: [gitty](https://aur.archlinux.org/packages/gitty/)
- Nix: `nix-env -iA nixpkgs.gitty`
- [Packages](https://github.com/muesli/gitty/releases) in Debian & RPM formats

### Binaries
- [Binaries](https://github.com/muesli/gitty/releases) for Linux, FreeBSD, OpenBSD, macOS, Windows

### From source

Make sure you have a working Go environment (Go 1.14 or higher is required).
See the [install instructions](http://golang.org/doc/install.html).

Compiling gitty is easy, simply run:

    git clone https://github.com/muesli/gitty.git
    cd gitty
    go build

## Usage

### Access Tokens

Note: In order to access the APIs of hosting providers like GitHub or GitLab,
`gitty` requires you to provide valid access tokens in an environment variable
called `GITTY_TOKENS`.

You can provide tokens for multiple hosts and services in this format:

`github.com=abc123;gitlab.com=xyz890;myhost.tld=...`

#### GitHub

You can [create a new token](https://github.com/settings/tokens/new?scopes=repo:status,public_repo,read:user,read:org&description=gitty)
in your profile settings:
Developer settings > Personal access tokens > Generate new token

Make sure to enable the `repo:status`, `public_repo`, `read:user`, and
`read:org` permissions in particular.

#### GitLab

You can create a new token in your profile settings:
User Settings > Access Tokens

Make sure to enable the `read_user`, `read_api`, and `read_repository`
permissions.

### Basic usage

You can start `gitty` with either a path or a URL as an argument. If no argument
was provided, `gitty` will operate on the current working directory.

```bash
$ gitty /some/repo
$ gitty github.com/some/project
$ gitty https://myhost.tld/some/project
```

The following flags are supported:

```
  -max-branch-age int
        Max age of a branch in days to be considered active (default 28)
  -max-branches int
        Max amount of active branches to show (default 10)
  -max-commits int
        Max amount of commits to show (default 10)
  -max-issues int
        Max amount of issues to show (default 10)
  -max-pull-requests int
        Max amount of pull requests to show (default 10)
```

### Open issue or pull request in browser

If you launch `gitty` with the ID of an issue or pull request, it will open the
issue or pull request in your browser:

```bash
gitty [PATH|URL] 42
```
