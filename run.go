package spicy

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(r io.Reader, args []string) (io.Reader, error)
}

type ExecRunner struct {
	command string
}

func NewRunner(cmd string) ExecRunner {
	return ExecRunner{command: cmd}
}

func (e ExecRunner) Run(r io.Reader, args []string) (io.Reader, error) {
	log.Infof("About to run %s %s\n", e.command, strings.Join(args, " "))
	cmd := exec.Command(e.command, args...)
	var out bytes.Buffer
	var errout bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errout
	err := cmd.Run()
	log.Debug("stdout: ", out.String())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error running '%s': %s", e.command, errout.String()))
	}
	return &out, nil
}
