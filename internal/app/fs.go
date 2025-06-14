package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type FileSystem interface {
	Exists(path string) (bool, error)
	Open(path, editor string) error
	Remove(path string) error
	GetGitRepos(dir string) ([]string, error)
}

type OSFileSystem struct {
	cmd CMD
}

func (f OSFileSystem) Exists(path string) (bool, error) {
	log.Printf("checking whether \"%s\" exists", path)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check whether \"%s\" exists: %s", path, err)
}
func (f OSFileSystem) Open(path, editor string) error {
	log.Printf("opening \"%s\" with \"%s\" editor", path, editor)
	_, err := f.cmd.ShellRun(editor, []string{path})
	if err != nil {
		return fmt.Errorf("failed to open \"%s\" with \"%s\" editor: %s", path, editor, err)
	}

	return nil
}
func (f OSFileSystem) Remove(path string) error {
	log.Printf("removing \"%s\"", path)
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove \"%s\": %s", path, err)
	}
	return nil
}
func (f OSFileSystem) GetGitRepos(dir string) ([]string, error) {
	log.Printf("gathering GIT directories from \"%s\"", dir)
	dirs := []string{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get directories from \"%s\": %s", dir, err)
	}

	for _, entry := range entries {
		if f.isGitRepo(filepath.Join(dir, entry.Name())) {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

func (f OSFileSystem) isGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

func NewOSFileSystem(cmd CMD) OSFileSystem {
	return OSFileSystem{
		cmd: cmd,
	}
}
