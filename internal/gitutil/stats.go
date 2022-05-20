package gitutil

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/cli/safeexec"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
)

var (
	// ErrGitFatal is returned by Git for fatal application errors, such as
	// if Git cannot find the .git directory.
	ErrGitFatal = errors.New("git error")

	// ErrGitUsage is returned by Git for errors in command line usage..
	ErrGitUsage = errors.New("git invalid usage")
)

// Stats contains info about a Git repository.
type Stats struct {
	CurrentBranch      string
	CurrentBranchSafe  string
	LatestTag          string
	Tags               []string
	CommitHash         string
	CommitShortHash    string
	CommitSubject      string
	CommitComitterDate string
	CommitAuthorDate   string
	Revision           int
	Remotes            map[string]Remote
	EstimatedRepoGroup string
	EstimatedRepoName  string
}

// Lookup tries to get a value based on the correlated built-in variable name.
// This method implements the varsub.Source interface.
//
// The string name -> field mapping is based on the documentation:
// https://iver-wharf.github.io/#/usage-wharfyml/variables/built-in-variables
func (s Stats) Lookup(name string) (varsub.Var, bool) {
	var value any
	switch name {
	case "GIT_BRANCH", "REPO_BRANCH":
		value = s.CurrentBranch
	case "GIT_COMMIT":
		value = s.CommitHash
	case "GIT_COMMIT_AUTHOR_DATE":
		value = s.CommitAuthorDate
	case "GIT_COMMIT_COMMITTER_DATE":
		value = s.CommitAuthorDate
	case "GIT_COMMIT_SUBJECT":
		value = s.CommitSubject
	case "GIT_SAFEBRANCH":
		value = s.CurrentBranchSafe
	case "GIT_TAG":
		value = s.LatestTag
	case "REPO_NAME":
		if s.EstimatedRepoName == "" {
			return varsub.Var{}, false
		}
		value = s.EstimatedRepoName
	case "REPO_GROUP":
		if s.EstimatedRepoGroup == "" {
			return varsub.Var{}, false
		}
		value = s.EstimatedRepoGroup
	default:
		return varsub.Var{}, false
	}
	return varsub.Var{
		Key:         name,
		Value:       value,
		SourceLabel: "git",
	}, true
}

var statsFields = []string{
	"GIT_BRANCH",
	"REPO_BRANCH",
	"GIT_COMMIT",
	"GIT_COMMIT_AUTHOR_DATE",
	"GIT_COMMIT_COMMITTER_DATE",
	"GIT_COMMIT_SUBJECT",
	"GIT_SAFEBRANCH",
	"GIT_TAG",
	"REPO_NAME",
	"REPO_GROUP",
}

// ListVars will return a slice of all variables that this varsub Source
// provides.
func (s Stats) ListVars() []varsub.Var {
	var vars []varsub.Var
	for _, key := range statsFields {
		value, ok := s.Lookup(key)
		if !ok {
			continue
		}
		vars = append(vars, value)
	}
	return vars
}

func (s Stats) String() string {
	var sb strings.Builder
	sb.WriteString("GIT_BRANCH=")
	sb.WriteString(s.CurrentBranch)
	sb.WriteString("\nGIT_COMMIT=")
	sb.WriteString(s.CommitHash)
	sb.WriteString("\nGIT_COMMIT_AUTHOR_DATE=")
	sb.WriteString(s.CommitAuthorDate)
	sb.WriteString("\nGIT_COMMIT_COMMITTER_DATE=")
	sb.WriteString(s.CommitComitterDate)
	sb.WriteString("\nGIT_COMMIT_SUBJECT=")
	sb.WriteString(s.CommitSubject)
	sb.WriteString("\nGIT_TAG=")
	sb.WriteString(s.LatestTag)
	sb.WriteString("\nREPO_GROUP=")
	sb.WriteString(s.EstimatedRepoGroup)
	sb.WriteString("\nREPO_NAME=")
	sb.WriteString(s.EstimatedRepoName)
	return sb.String()
}

// Remote is a Git remote, containing the fetch and pull URLs.
type Remote struct {
	FetchURL string
	PushURL  string
}

// StatsFromExec obtains Git repo stats by executing different Git commands.
func StatsFromExec(dir string) (Stats, error) {
	currentBranch, err := execGitCmd(dir, "branch", "--show-current")
	if err != nil {
		return Stats{}, err
	}

	// %n: newline
	// %H: long SHA hash
	// %h: short SHA hash
	// %s: oneline summary from commit message
	// %aI: author date (ISO 8601 formatted)
	// %cI: committer date (ISO 8601 formatted)
	commitInfo, err := execGitCmdLines(dir,
		"log", "-n", "1", "HEAD", "--format=%H%n%h%n%s%n%aI%n%cI")
	if err != nil {
		return Stats{}, err
	}

	revisionStr, err := execGitCmd(dir, "rev-list", "--count", "HEAD")
	if err != nil {
		return Stats{}, err
	}
	revision, err := strconv.ParseInt(revisionStr, 10, 0)
	if err != nil {
		return Stats{}, err
	}

	tags, err := execGitCmdLines(dir, "tag", "--sort=-taggerdate", "--points-at")
	if err != nil {
		return Stats{}, err
	}

	remotesStrs, err := execGitCmdLines(dir, "remote", "--verbose", "show", "-n")
	if err != nil {
		return Stats{}, err
	}
	remotes := parseRemotes(remotesStrs)

	stats := Stats{
		CurrentBranch:      currentBranch,
		CurrentBranchSafe:  strings.ReplaceAll(currentBranch, "/", "-"),
		CommitHash:         safeGetTrimmed(commitInfo, 0),
		CommitShortHash:    safeGetTrimmed(commitInfo, 1),
		CommitSubject:      safeGetTrimmed(commitInfo, 2),
		CommitAuthorDate:   safeGetTrimmed(commitInfo, 3),
		CommitComitterDate: safeGetTrimmed(commitInfo, 4),
		Revision:           int(revision),
		Tags:               tags,
		LatestTag:          safeGetTrimmed(tags, 0),
		Remotes:            remotes,
	}

	for name, r := range remotes {
		if name != "origin" {
			continue
		}
		stats.EstimatedRepoGroup, stats.EstimatedRepoName = estimateRepoGroupAndName(r)
		break
	}

	return stats, nil
}

func safeGetTrimmed(slice []string, index int) string {
	if index >= len(slice) {
		return ""
	}
	return strings.TrimSpace(slice[index])
}

func parseRemotes(strs []string) map[string]Remote {
	remotes := make(map[string]Remote)
	for _, line := range strs {
		var name, url, kind string
		_, err := fmt.Sscanf(line, "%s\t%s %s", &name, &url, &kind)
		if err != nil {
			continue
		}
		r := remotes[name]
		switch kind {
		case "(fetch)":
			r.FetchURL = url
		case "(push)":
			r.PushURL = url
		}
		remotes[name] = r
	}
	return remotes
}

func execGitCmdLines(dir string, args ...string) ([]string, error) {
	output, err := execGitCmd(dir, args...)
	if err != nil {
		return nil, err
	}
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}
	return strings.Split(output, "\n"), nil
}

func execGitCmd(dir string, args ...string) (string, error) {
	gitBin, err := safeexec.LookPath("git")
	if err != nil {
		return "", wrapGitExecError(err, args)
	}
	cmd := exec.Command(gitBin, append([]string{"-C", dir, "--no-pager"}, args...)...)
	outBytes, err := cmd.CombinedOutput()
	outBytes = bytes.TrimSpace(outBytes)
	if err != nil {
		return "", convGitExecError(err, outBytes, args)
	}
	return string(outBytes), nil
}

func convGitExecError(err error, outBytes []byte, args []string) error {
	_, isExecError := err.(*exec.Error)
	if isExecError {
		// No need to wrap it. The exec error contains enough context.
		return err
	}
	exitErr, isExitError := err.(*exec.ExitError)
	if !isExitError {
		return wrapGitExecError(err, args)
	}
	// https://git-scm.com/docs/api-error-handling
	switch exitErr.ExitCode() {
	case 128:
		return wrapGitExecError(fmt.Errorf("%w: %s", ErrGitFatal, outBytes), args)
	case 129:
		return wrapGitExecError(fmt.Errorf("%w: %s", ErrGitUsage, outBytes), args)
	default:
		return wrapGitExecError(err, args)
	}
}

func wrapGitExecError(err error, args []string) error {
	return fmt.Errorf("exec %q: %w",
		strings.Join(append([]string{"git"}, args...), " "), err)
}

// Regex patterns for estimating the repo name and group names. One hidden
// gem is the (?:v\d+/)? part that removes any versioned paths, ex the "/v3/",
// that Azure DevOps uses.
var estURLRegex = regexp.MustCompile(
	`\w+://[^/]+/(?:v\d+/)?(.*)/([^/]+)`)
var estSSHRegex = regexp.MustCompile(
	`\w+:(?:v\d+/)?(.*)/([^/]+)`)

func estimateRepoGroupAndName(origin Remote) (string, string) {
	url := origin.FetchURL
	if url == "" {
		url = origin.PushURL
	}
	if url == "" {
		return "", ""
	}
	groups := estURLRegex.FindStringSubmatch(origin.FetchURL)
	if groups == nil {
		groups = estSSHRegex.FindStringSubmatch(origin.FetchURL)
	}
	if groups == nil {
		return "", ""
	}
	// Typical of Azure DevOps to have a trailing /_git in the path
	return strings.TrimSuffix(groups[1], "/_git"),
		strings.TrimSuffix(groups[2], ".git")
}
