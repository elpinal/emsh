package command

import (
	"os"
	"os/exec"
)

type Cmd interface {
	Run() error
}

type OsCmd exec.Cmd

type Exit struct {
	Code int
}

func (e Exit) Run() error {
	os.Exit(e.Code)
	return nil
}
