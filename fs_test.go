package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
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

func assertNotExists(t *testing.T, dir string, exists bool) {
	if exists {
		t.Fatalf("expected %s to not exist, but it exists", dir)
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

func assertHistory(t *testing.T, obj any, expected []map[string]any) {
	var his []map[string]any
	var objType string

	if o, ok := obj.(*FakeCMD); ok {
		his = o.history
		objType = "CMD"
	}

	if !reflect.DeepEqual(his, expected) {
		if len(his) == len(expected) && len(his) == 0 {
			return
		}
		t.Fatalf("invalid %s history. Expected %v, got %v", objType, expected, his)
	}
}

func TestExists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		cmd := &FakeCMD{}
		fs := NewOSFileSystem(cmd)

		dir := t.TempDir()
		exists, err := fs.Exists(dir)
		assertNotError(t, err)
		if !exists {
			t.Fatalf("expected %s to exist, but it does not", dir)
		}

		assertHistory(t, cmd, []map[string]any{})
	})
	t.Run("does not exist", func(t *testing.T) {
		cmd := &FakeCMD{}
		fs := NewOSFileSystem(cmd)

		dir := t.TempDir()
		exists, err := fs.Exists(filepath.Join(dir, "any"))
		assertNotError(t, err)
		assertNotExists(t, dir, exists)
	})
}

func TestOpen(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cmd := &FakeCMD{}
		fs := NewOSFileSystem(cmd)

		editor := "any_editor"
		dir := t.TempDir()
		err := fs.Open(dir, editor)
		assertNotError(t, err)
		assertHistory(t, cmd, []map[string]any{
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
		assertError(t, err)
	})
}

func TestRemove(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cmd := &FakeCMD{}
		fs := NewOSFileSystem(cmd)

		dir := t.TempDir()
		err := fs.Remove(dir)
		assertNotError(t, err)
		assertHistory(t, cmd, []map[string]any{})

		exists, _ := fs.Exists(dir)
		assertNotExists(t, dir, exists)
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
		assertNotError(t, err)
		assertHistory(t, cmd, []map[string]any{})

		expectedDirs := []string{"one", "two"}
		if !reflect.DeepEqual(dirs, expectedDirs) {
			t.Fatalf("expected dirs %v, got %v", expectedDirs, dirs)
		}
	})
}
