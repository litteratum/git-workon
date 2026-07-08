package app

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/kirsle/configdir"
)

// projectsPath lives in the config directory (next to config.json) rather than
// the cache directory: the registry is durable, not disposable. The configdir
// library has no dedicated data-dir helper, so we reuse ConfigDir for
// consistency with the rest of the configuration handling.
var projectsPath = filepath.Join(ConfigDir, "projects.json")

// IProjects is a durable registry of known project names.
//
// Unlike the cache, it is never treated as disposable: it survives cache
// removal and is the source of truth for shell completion. The cache maps a
// project to the source it was cloned from and may be cleared at any time; this
// registry only records that a project has been seen.
type IProjects interface {
	Add(project string)
	List() []string
}

type Projects struct {
	Names []string
}

func (p *Projects) List() []string {
	return p.Names
}

// Add records a project name and persists the registry. It is a no-op if the
// name is already known.
func (p *Projects) Add(project string) {
	if slices.Contains(p.Names, project) {
		return
	}
	p.Names = append(p.Names, project)
	slices.Sort(p.Names)
	p.write()
}

func (p *Projects) write() {
	data, err := json.MarshalIndent(p.Names, "", "  ")
	if err != nil {
		log.Printf("failed to marshal the projects registry: %s", err)
		return
	}

	if err = os.WriteFile(projectsPath, data, 0644); err != nil {
		log.Printf("failed to write the projects registry at %s: %s", projectsPath, err)
	}
}

func NewProjects(names []string) *Projects {
	return &Projects{Names: names}
}

func NewProjectsFromFile() *Projects {
	data, err := os.ReadFile(projectsPath)
	if os.IsNotExist(err) {
		return createProjectsFile()
	}

	if err != nil {
		log.Fatalf("failed to read the projects registry at %s: %s", projectsPath, err)
	}

	var names []string
	if err = json.Unmarshal(data, &names); err != nil {
		log.Fatalf("failed to unmarshal the projects registry at %s: %s", projectsPath, err)
	}

	return NewProjects(names)
}

func createProjectsFile() *Projects {
	if err := configdir.MakePath(ConfigDir); err != nil {
		log.Printf("failed to create config dir %s: %s", ConfigDir, err)
		return NewProjects(nil)
	}

	projects := NewProjects([]string{})
	projects.write()
	return projects
}
