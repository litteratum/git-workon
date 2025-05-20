package main

import (
	"os"
	"path/filepath"
	"testing"
)

func createFile(t *testing.T, name string) string {
	path := filepath.Join(t.TempDir(), name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create a test file: %s", err)
	}
	f.Close()
	return path
}

func assertError(t *testing.T, err error) {
	if err == nil {
		t.Fatal("expected error but did not get one")
	}
}

func TestOSExecRun(t *testing.T) {
	t.Run("rm file; file exists; removed", func(t *testing.T) {
		exec := OSExec{}
		path := createFile(t, "filename")
		_, err := exec.Run("rm", []string{path})
		assertNotError(t, err)
	})
	t.Run("rm file; file does not exist; error", func(t *testing.T) {
		exec := OSExec{}
		dir := t.TempDir()
		_, err := exec.Run("rm", []string{filepath.Join(dir, "any.txt")})
		assertError(t, err)
	})
}

func TestOSExecRunCwd(t *testing.T) {
	t.Run("rm file; file exists; removed", func(t *testing.T) {
		exec := OSExec{}
		dir := t.TempDir()
		path := createFile(t, "filename")
		_, err := exec.RunCwd(dir, "rm", []string{path})
		assertNotError(t, err)
	})
}
