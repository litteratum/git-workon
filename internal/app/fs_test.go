package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type FakeCMD struct {
	history []map[string]any
	err     error
	results []CMDResult
}

func (fc *FakeCMD) Run(name string, args []string) (CMDResult, error) {
	fc.history = append(
		fc.history,
		map[string]any{
			"_method": "Run",
			"name":    name,
			"args":    args,
		},
	)

	return fc.getNextResult(), fc.err
}

func (fc *FakeCMD) RunCwd(dir string, name string, args []string) (CMDResult, error) {
	fc.history = append(
		fc.history,
		map[string]any{
			"_method": "RunCwd",
			"dir":     dir,
			"name":    name,
			"args":    args,
		},
	)

	return fc.getNextResult(), fc.err
}

func (fc *FakeCMD) ShellRun(name string, args []string) (CMDResult, error) {
	fc.history = append(
		fc.history,
		map[string]any{
			"_method": "ShellRun",
			"name":    name,
			"args":    args,
		},
	)

	return fc.getNextResult(), fc.err
}

func (fc *FakeCMD) getNextResult() CMDResult {
	if len(fc.results) > 0 {
		result := fc.results[0]
		fc.results = fc.results[1:]
		return result
	} else {
		return CMDResult{}
	}
}

func createGitDir(t *testing.T, root, name string) string {
	path := filepath.Join(root, name)
	err := os.MkdirAll(filepath.Join(path, ".git"), 0755)
	if err != nil {
		t.Fatalf("failed to create a testing dir \"%s\": %s", path, err)
	}
	return path
}

func TestExists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		cmd := &FakeCMD{}
		fs := NewOSFileSystem(cmd)
		dir := t.TempDir()

		exists, err := fs.Exists(dir)

		require.NoError(t, err)
		require.True(t, exists, "must exist")
		require.Empty(t, cmd.history)
	})
	t.Run("does not exist", func(t *testing.T) {
		cmd := &FakeCMD{}
		fs := NewOSFileSystem(cmd)
		dir := t.TempDir()

		exists, err := fs.Exists(filepath.Join(dir, "any"))

		require.NoError(t, err)
		require.False(t, exists, "must not exist")
	})
}

func TestOpen(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cmd := &FakeCMD{}
		fs := NewOSFileSystem(cmd)
		editor := "any_editor"
		dir := t.TempDir()

		err := fs.Open(dir, editor)

		require.NoError(t, err)
		require.Equal(t, cmd.history, []map[string]any{
			{
				"_method": "ShellRun",
				"name":    editor,
				"args":    []string{dir},
			},
		})
	})
	t.Run("error", func(t *testing.T) {
		cmd := &FakeCMD{
			err: errors.New("any err"),
		}
		fs := NewOSFileSystem(cmd)
		editor := "any_editor"
		dir := t.TempDir()

		err := fs.Open(dir, editor)

		require.Error(t, err)
	})
}

func TestRemove(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cmd := &FakeCMD{}
		fs := NewOSFileSystem(cmd)
		dir := t.TempDir()
		err := fs.Remove(dir)
		require.NoError(t, err)
		require.Empty(t, cmd.history)

		exists, err := fs.Exists(dir)
		require.NoError(t, err)
		require.False(t, exists, "must not exist after removed")
	})
}
func TestGetGitRepos(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cmd := &FakeCMD{}
		fs := NewOSFileSystem(cmd)

		dir := t.TempDir()
		createGitDir(t, dir, "one")
		gitDir2 := createGitDir(t, dir, "two")
		createGitDir(t, gitDir2, "three")

		dirs, err := fs.GetGitRepos(dir)
		require.NoError(t, err)
		require.Empty(t, cmd.history)
		require.Equal(t, dirs, []string{"one", "two"})
	})
}
