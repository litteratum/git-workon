package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type CMDResult struct {
	Stdout string
	Stderr string
}

type CMD interface {
	Run(name string, args []string) (CMDResult, error)
	RunCwd(dir string, name string, args []string) (CMDResult, error)
	ShellRun(name string, args []string) (CMDResult, error)
}

type OSExec struct{}

func (ose OSExec) Run(name string, args []string) (CMDResult, error) {
	log.Printf("executing \"%s\" with args %s", name, args)
	cmd := exec.Command(name, args...)
	return ose.run(cmd)
}

func (ose OSExec) RunCwd(dir string, name string, args []string) (CMDResult, error) {
	log.Printf("executing \"%s\" with args %s in \"%s\"", name, args, dir)
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return ose.run(cmd)
}

func (ose OSExec) ShellRun(name string, args []string) (CMDResult, error) {
	log.Printf("executing \"%s\" with args %s", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	cmdResult := CMDResult{}
	if err != nil {
		return cmdResult, err
	}

	return cmdResult, nil
}

func (ose OSExec) run(cmd *exec.Cmd) (CMDResult, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	cmdResult := CMDResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err != nil {
		stderrContent := stderr.String()
		var message string
		if stderrContent != "" {
			message = fmt.Sprintf("%s.\n%s", err, stderrContent)
		} else {
			message = err.Error()
		}
		return cmdResult, errors.New(message)
	}

	return cmdResult, nil
}

func NewOSExec() OSExec {
	return OSExec{}
}
