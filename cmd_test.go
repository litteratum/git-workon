package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestOSExecRun(t *testing.T) {
	t.Run("rm file; file exists; removed", func(t *testing.T) {
		exec := OSExec{}
		path := createFile(t, "filename")
		_, err := exec.Run("rm", []string{path})
		require.NoError(t, err)
	})
	t.Run("rm file; file does not exist; error", func(t *testing.T) {
		exec := OSExec{}
		dir := t.TempDir()
		_, err := exec.Run("rm", []string{filepath.Join(dir, "any.txt")})
		require.Error(t, err)
	})
}

func TestOSExecRunCwd(t *testing.T) {
	t.Run("rm file; file exists; removed", func(t *testing.T) {
		exec := OSExec{}
		dir := t.TempDir()
		path := createFile(t, "filename")
		_, err := exec.RunCwd(dir, "rm", []string{path})
		require.NoError(t, err)
	})
}
