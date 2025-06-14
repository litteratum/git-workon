package app

import (
	"fmt"
	"log"
	"strings"
)

type Git interface {
	GetProjectState(path string) (GitProjectState, error)
	Clone(source, destination string) error
}

type GitProjectState struct {
	Stashes string
	Tags    string
	Commits string
	Status  string
}

func (state GitProjectState) String() string {
	output := strings.Builder{}
	if state.Stashes != "" {
		output.WriteString(fmt.Sprintf("Stashes:\n%s", state.Stashes))
	}
	if state.Tags != "" {
		output.WriteString(fmt.Sprintf("\nTags:\n%s", state.Tags))
	}
	if state.Commits != "" {
		output.WriteString(fmt.Sprintf("\nCommits:\n%s", state.Commits))
	}
	if state.Status != "" {
		output.WriteString(fmt.Sprintf("\nStatus:\n%s", state.Status))
	}
	return output.String()
}

func (state GitProjectState) Clean() bool {
	fields := []string{state.Stashes, state.Tags, state.Commits, state.Status}
	for _, field := range fields {
		if field != "" {
			return false
		}
	}
	return true
}

type GitAPI struct {
	cmd CMD
}

func (g GitAPI) GetProjectState(path string) (GitProjectState, error) {
	log.Printf("getting Git status for \"%s\"", path)
	stashes, err := g.getGitStashes(path)
	if err != nil {
		return GitProjectState{}, fmt.Errorf("failed to get stashes for \"%s\": %s", path, err)
	}
	tags, err := g.getGitTags(path)
	if err != nil {
		return GitProjectState{}, fmt.Errorf("failed to get tags for \"%s\": %s", path, err)
	}
	commits, err := g.getGitCommits(path)
	if err != nil {
		return GitProjectState{}, fmt.Errorf("failed to get commits for \"%s\": %s", path, err)
	}
	status, err := g.getGitStatus(path)
	if err != nil {
		return GitProjectState{}, fmt.Errorf("failed to get status for \"%s\": %s", path, err)
	}

	return GitProjectState{
		Stashes: stashes,
		Tags:    tags,
		Commits: commits,
		Status:  status,
	}, nil
}

func (g GitAPI) Clone(source, destination string) error {
	log.Printf("cloning \"%s\" to \"%s\"", source, destination)
	_, err := g.cmd.Run(
		"git",
		[]string{
			"clone",
			source,
			destination,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to clone \"%s\" to \"%s\": %s", source, destination, err)
	}

	return nil
}

func (g GitAPI) getGitStashes(path string) (string, error) {
	result, err := g.cmd.RunCwd(path, "git", []string{"stash", "list"})
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}

func (g GitAPI) getGitTags(path string) (string, error) {
	result, err := g.cmd.RunCwd(path, "git", []string{"push", "--tags", "--dry-run"})
	if err != nil {
		return "", err
	}
	if !strings.Contains(result.Stderr, "new tag") {
		return "", nil
	}
	return result.Stderr, nil
}

func (g GitAPI) getGitCommits(path string) (string, error) {
	result, err := g.cmd.RunCwd(
		path,
		"git",
		[]string{"log", "--branches", "--not", "--remotes", "--decorate", "--oneline"},
	)
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}

func (g GitAPI) getGitStatus(path string) (string, error) {
	result, err := g.cmd.RunCwd(path, "git", []string{"status", "--short"})
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}

func NewGitAPI(cmd CMD) GitAPI {
	return GitAPI{
		cmd: cmd,
	}
}
