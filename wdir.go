package main

import (
	"fmt"
	"log"
	"path"
	"sync"
)

type WorkingDir struct {
	directory string
	git       Git
	fs        FileSystem
}

func NewWorkingDir(directory string) WorkingDir {
	cmd := NewOSExec()
	return WorkingDir{
		directory: directory,
		git:       NewGitAPI(cmd),
		fs:        NewOSFileSystem(cmd),
	}
}

type StartOpts struct {
	Open bool
}

type DoneOpts struct {
	Force bool
}

func (wd WorkingDir) Start(projects []string, sources, editors []string, opts StartOpts) error {
	if len(projects) == 0 {
		return fmt.Errorf("no projects to start specified")
	}
	if len(sources) == 0 {
		return fmt.Errorf("no GIT sources specified")
	}

	var lastProjectPath string

	for _, project := range projects {
		projPath := wd.projectPath(project)
		exists, err := wd.fs.Exists(projPath)
		if err != nil {
			log.Printf("failed to check whether \"%s\" exists: %s", project, err)
			continue
		}
		if exists {
			lastProjectPath = projPath
			log.Printf("\"%s\" already exists. No need to clone", project)
			continue
		}

		err = wd.clone(wd.directory, project, sources)
		if err != nil {
			log.Println(err)
			continue
		}

		lastProjectPath = projPath
	}

	if lastProjectPath == "" {
		return fmt.Errorf("failed to start any project")
	}

	if opts.Open {
		err := wd.open(lastProjectPath, editors)
		if err != nil {
			return err
		}
	}

	return nil
}

func (wd WorkingDir) Done(projects []string, opts DoneOpts) error {
	gitRepos := []string{}
	if len(projects) > 0 {
		gitRepos = append(gitRepos, projects...)
	} else {
		var err error
		gitRepos, err = wd.fs.GetGitRepos(wd.directory)
		if err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(gitRepos))
	for _, repo := range gitRepos {
		go func(r string) {
			defer wg.Done()
			wd.done(repo, opts)
		}(repo)
	}

	wg.Wait()
	return nil
}

func (wd WorkingDir) clone(dir, project string, sources []string) error {
	for _, source := range sources {
		err := wd.git.Clone(path.Join(source, project), wd.projectPath(project))
		if err != nil {
			log.Printf("%s\nTrying other sources...", err)
		} else {
			return nil
		}
	}

	return fmt.Errorf("failed to clone \"%s\". Tried all configured sources", project)
}

func (wd WorkingDir) open(path string, editors []string) error {
	for _, editor_ := range editors {
		err := wd.fs.Open(path, editor_)
		if err != nil {
			log.Printf("%s. Will try other editors", err)
		} else {
			return nil
		}
	}

	return fmt.Errorf("failed to open \"%s\". Tried all configured editors", path)
}

func (wd WorkingDir) done(project string, opts DoneOpts) {
	projectPath := wd.projectPath(project)

	if opts.Force {
		log.Printf("forcefully removing \"%s\"", projectPath)
		wd.removeSafe(projectPath)
		return
	}

	state, err := wd.git.GetProjectState(projectPath)
	if err != nil {
		log.Printf("failed to get state of \"%s\": %s", projectPath, err)
		return
	}

	if state.Clean() {
		wd.removeSafe(projectPath)
	} else {
		log.Printf("\"%s\" will not be removed: the project is not clean:\n%s", projectPath, state)
	}
}

func (wd WorkingDir) removeSafe(path string) {
	err := wd.fs.Remove(path)
	if err != nil {
		log.Printf("failed to remove \"%s\": %s", path, err)
	}
}

func (wd WorkingDir) projectPath(name string) string {
	return path.Join(wd.directory, name)
}
