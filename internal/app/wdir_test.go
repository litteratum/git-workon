package app

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type wdComponents struct {
	dir   string
	fs    FileSystem
	git   Git
	cache ICache
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

type FakeCache struct {
	Cache
	Writes int
}

func (fc *FakeCache) Write() {
	fc.Writes++
}

func NewEmptyFakeCache() *FakeCache {
	return &FakeCache{
		Cache: Cache{
			Data: map[string]ProjectInfo{},
		},
	}
}
func NewFakeCache(data map[string]ProjectInfo) *FakeCache {
	return &FakeCache{
		Cache: Cache{
			Data: data,
		},
	}
}

func buildWorkingDir(comps wdComponents) WorkingDir {
	if comps.dir == "" {
		comps.dir = "/dwd"
	}
	if comps.cache == nil {
		comps.cache = NewEmptyFakeCache()
	}
	return WorkingDir{
		directory: comps.dir,
		fs:        comps.fs,
		git:       comps.git,
		config:    NewDefaultConfig(),
		cache:     comps.cache,
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
		err := wd.Start(
			[]string{"proj"},
			[]string{"dsource", "dsource2"},
			"vi",
			StartOpts{Open: false},
		)

		require.NoError(t, err)
		repo, ok := fs.repos["/dwd/proj"]
		require.True(t, ok)
		require.Equal(t, repo.opensCount, 0)
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
		err := wd.Start([]string{"proj"}, []string{"dsource"}, "vim", StartOpts{Open: true})
		require.NoError(t, err)

		repo := fs.repos["/dwd/proj"]
		require.Equal(t, repo.opensCount, 1)
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
		err := wd.Start([]string{"proj"}, []string{"dsource"}, "vim", StartOpts{Open: true})
		require.NoError(t, err)

		repo := fs.repos["/dwd/proj"]
		require.Equal(t, repo.opensCount, 1)
	})
	t.Run("projects are empty", func(t *testing.T) {
		wd := buildWorkingDir(wdComponents{})
		err := wd.Start([]string{}, []string{"s1"}, "", StartOpts{})
		require.Error(t, err)
	})
	t.Run("sources are empty", func(t *testing.T) {
		wd := buildWorkingDir(wdComponents{})
		err := wd.Start([]string{"p1"}, []string{}, "", StartOpts{})
		require.Error(t, err)
	})
	t.Run("source from cache", func(t *testing.T) {
		fs := NewFakeFS().WithEditors(
			map[string]FakeEditor{"vim": {}},
		)
		git := NewFakeGit(fs).WithSources([]string{"sc/p1", "s/p1"})
		cache := NewFakeCache(
			map[string]ProjectInfo{
				"p1": {
					Source: "sc",
				},
			},
		)
		wd := buildWorkingDir(wdComponents{
			fs:    fs,
			git:   git,
			cache: cache,
		})

		err := wd.Start([]string{"p1"}, []string{}, "", StartOpts{})
		require.NoError(t, err)
		require.Equal(t, cache.Writes, 1)
	})
	t.Run("cloned source cached", func(t *testing.T) {
		fs := NewFakeFS()
		git := NewFakeGit(fs).WithSources([]string{"s/p1"})
		cache := NewFakeCache(map[string]ProjectInfo{})
		wd := buildWorkingDir(wdComponents{
			fs:    fs,
			git:   git,
			cache: cache,
		})

		err := wd.Start([]string{"p1"}, []string{"s"}, "", StartOpts{})
		require.NoError(t, err)
		require.Equal(t, cache.Data, map[string]ProjectInfo{"p1": {Source: "s"}})
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
					require.NoError(t, err)
					require.Len(t, fs.repos, 0)
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
		require.NoError(t, err)
		require.Len(t, fs.repos, 0)
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
		require.NoError(t, err)
		require.Len(t, fs.repos, 1)
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
		require.NoError(t, err)
		require.Len(t, fs.repos, 0)
	})
}
