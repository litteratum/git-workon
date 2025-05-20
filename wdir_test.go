package main

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"testing"
)

type wdComponents struct {
	dir string
	fs  FileSystem
	git Git
}

type FakeRepo struct {
	path       string
	opensCount int
}

type FakeEditor struct{}

func (f FakeEditor) Open(repo *FakeRepo) error {
	repo.opensCount++
	return nil
}

type FakeFS struct {
	repos   map[string]*FakeRepo
	editors map[string]FakeEditor
}

func NewFakeFS() *FakeFS {
	return &FakeFS{
		repos:   map[string]*FakeRepo{},
		editors: map[string]FakeEditor{"vi": {}},
	}
}

func (f FakeFS) WithRepos(repos map[string]*FakeRepo) *FakeFS {
	f.repos = repos
	return &f
}

func (f FakeFS) WithEditors(editors map[string]FakeEditor) *FakeFS {
	f.editors = editors
	return &f
}

func (f *FakeFS) Exists(path string) (bool, error) {
	if _, ok := f.repos[path]; ok {
		return true, nil
	}
	return false, nil
}
func (f *FakeFS) Open(path, editor string) error {
	e, ok := f.editors[editor]
	if !ok {
		return fmt.Errorf("unknown editor: %s", editor)
	}
	for _, repo := range f.repos {
		if repo.path == path {
			return e.Open(repo)
		}
	}
	return fmt.Errorf("unknown repo: %s", path)
}
func (f *FakeFS) Remove(path string) error {
	delete(f.repos, path)
	return nil
}
func (f *FakeFS) GetGitRepos(dir string) ([]string, error) {
	repos := []string{}
	for _, repo := range f.repos {
		if !strings.HasPrefix(repo.path, dir) {
			continue
		}
		repos = append(repos, path.Base(repo.path))
	}
	return repos, nil
}

type FakeGit struct {
	states  map[string]GitProjectState
	fs      FakeFS
	sources []string
}

func NewFakeGit(fs *FakeFS) *FakeGit {
	return &FakeGit{
		states: map[string]GitProjectState{},
		fs:     *fs,
	}
}

func (fg *FakeGit) WithStates(states map[string]GitProjectState) *FakeGit {
	fg.states = states
	return fg
}

func (fg *FakeGit) WithSources(sources []string) *FakeGit {
	fg.sources = sources
	return fg
}

func (fg *FakeGit) GetProjectState(path string) (GitProjectState, error) {
	if state, ok := fg.states[path]; ok {
		return state, nil
	}
	return GitProjectState{}, nil
}
func (fg *FakeGit) Clone(source, destination string) error {
	if !slices.Contains(fg.sources, source) {
		return fmt.Errorf("source \"%s\" not found", source)
	}
	fg.fs.repos[destination] = &FakeRepo{path: destination}
	return nil
}

func buildWorkingDir(comps wdComponents) WorkingDir {
	if comps.dir == "" {
		comps.dir = "/dwd"
	}
	return WorkingDir{
		directory: comps.dir,
		fs:        comps.fs,
		git:       comps.git,
	}
}

func assertNotError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("did not expect error, got %s", err)
	}
}

func TestStart(t *testing.T) {
	t.Run("no project; cloned; not opened", func(t *testing.T) {
		fs := NewFakeFS()
		git := NewFakeGit(fs).WithSources([]string{"dsource/proj", "dsource2/proj"})
		wd := buildWorkingDir(
			wdComponents{
				fs:  fs,
				git: git,
			},
		)
		err := wd.Start([]string{"proj"}, []string{"dsource", "dsource2"}, []string{"vi"}, StartOpts{Open: false})
		assertNotError(t, err)

		repo, ok := fs.repos["/dwd/proj"]
		if !ok {
			t.Fatalf("expected repo created")
		}
		if repo.opensCount != 0 {
			t.Fatalf("expected repo to not be opened")
		}
	})
	t.Run("no project; cloned; opened", func(t *testing.T) {
		fs := NewFakeFS().WithEditors(map[string]FakeEditor{"vim": {}})
		git := NewFakeGit(fs).WithSources([]string{"dsource/proj"})
		wd := buildWorkingDir(
			wdComponents{
				fs:  fs,
				git: git,
			},
		)
		err := wd.Start([]string{"proj"}, []string{"dsource"}, []string{"vim", "vi"}, StartOpts{Open: true})
		assertNotError(t, err)

		repo, ok := fs.repos["/dwd/proj"]
		if !ok {
			t.Fatalf("expected repo created")
		}
		if repo.opensCount != 1 {
			t.Fatalf("expected repo to be opened, got %d", repo.opensCount)
		}
	})
	t.Run("project exists; not cloned; opened", func(t *testing.T) {
		fs := NewFakeFS().WithEditors(
			map[string]FakeEditor{"vim": {}},
		).WithRepos(
			map[string]*FakeRepo{"/dwd/proj": {path: "/dwd/proj"}},
		)
		git := NewFakeGit(fs)
		wd := buildWorkingDir(
			wdComponents{
				fs:  fs,
				git: git,
			},
		)
		err := wd.Start([]string{"proj"}, []string{"dsource"}, []string{"vim", "vi"}, StartOpts{Open: true})
		assertNotError(t, err)

		repo := fs.repos["/dwd/proj"]
		if repo.opensCount != 1 {
			t.Fatalf("expected repo to be opened, got %d", repo.opensCount)
		}
	})
	t.Run("projects are empty", func(t *testing.T) {
		wd := buildWorkingDir(wdComponents{})
		err := wd.Start([]string{}, []string{"s1"}, []string{}, StartOpts{})
		assertError(t, err)
	})
	t.Run("sources are empty", func(t *testing.T) {
		wd := buildWorkingDir(wdComponents{})
		err := wd.Start([]string{"p1"}, []string{}, []string{}, StartOpts{})
		assertError(t, err)
	})
}

func TestDone(t *testing.T) {
	t.Run("specific projects; forcedly removed", func(t *testing.T) {
		tests := map[string]struct {
			gitProjectState GitProjectState
		}{
			"clean": {
				gitProjectState: GitProjectState{},
			},
			"dirty": {
				gitProjectState: GitProjectState{
					Stashes: "stashes",
				},
			},
		}

		for name, test := range tests {
			t.Run(
				name,
				func(t *testing.T) {
					fs := NewFakeFS()
					git := NewFakeGit(fs).WithStates(
						map[string]GitProjectState{
							"/dwd/proj": test.gitProjectState,
						},
					)
					wd := buildWorkingDir(
						wdComponents{
							fs:  fs,
							git: git,
						},
					)

					err := wd.Done([]string{"proj"}, DoneOpts{Force: true})
					assertNotError(t, err)

					if len(fs.repos) != 0 {
						t.Fatalf("expected all repos to be removed, got %d", len(fs.repos))
					}
				},
			)
		}
	})
	t.Run("specific projects; clean; removed", func(t *testing.T) {
		fs := NewFakeFS().WithRepos(
			map[string]*FakeRepo{
				"/dwd/proj":  {path: "/dwd/proj"},
				"/dwd/proj2": {path: "/dwd/proj2"},
			},
		)
		wd := buildWorkingDir(
			wdComponents{
				fs:  fs,
				git: NewFakeGit(fs),
			},
		)
		err := wd.Done([]string{"proj", "proj2"}, DoneOpts{})
		assertNotError(t, err)

		if len(fs.repos) != 0 {
			t.Fatalf("expected all repos to be removed, got %d", len(fs.repos))
		}
	})
	t.Run("specific projects; not clean; not removed", func(t *testing.T) {
		fs := NewFakeFS().WithRepos(
			map[string]*FakeRepo{
				"/dwd/proj": {path: "/dwd/proj"},
			},
		)
		git := NewFakeGit(fs).WithStates(
			map[string]GitProjectState{
				"/dwd/proj": {
					Status: "dirty",
				},
			},
		)
		wd := buildWorkingDir(
			wdComponents{
				fs:  fs,
				git: git,
			},
		)
		err := wd.Done([]string{"proj", "proj2"}, DoneOpts{})
		assertNotError(t, err)

		if len(fs.repos) != 1 {
			t.Fatalf("expected 1 repo to be left, got %d", len(fs.repos))
		}
	})
	t.Run("all projects", func(t *testing.T) {
		fs := NewFakeFS().WithRepos(
			map[string]*FakeRepo{
				"/dwd/proj":  {path: "/dwd/proj"},
				"/dwd/proj2": {path: "/dwd/proj2"},
			},
		)
		git := NewFakeGit(fs)
		wd := buildWorkingDir(
			wdComponents{
				fs:  fs,
				git: git,
			},
		)
		err := wd.Done([]string{}, DoneOpts{})
		assertNotError(t, err)

		if len(fs.repos) != 0 {
			t.Fatalf("expected all repos to be removed, got %d", len(fs.repos))
		}
	})
}
