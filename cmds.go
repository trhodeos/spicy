package spicy

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"strings"
)

func RunCmd(command string, args ...string) error {
	log.Infof("About to run %s %s\n", command, strings.Join(args, " "))
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	var errout bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errout
	err := cmd.Run()
	log.Debug(command, " stdout: ", out.String())
	if err != nil {
		log.Error("Error running ", command, ". Stderr output: ", errout.String())
	}
	return err
}

func RunCmdReturnStdout(command string, stdin io.Reader, args ...string) (io.Reader, error) {
	log.Infof("About to run %s %s\n", command, strings.Join(args, " "))
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	var errout bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errout
	cmd.Stdin = stdin
	err := cmd.Run()
	log.Debug(command, " stdout: ", out.String())
	if err != nil {
		log.Error("Error running ", command, ". Stderr output: ", errout.String())
	}
	return strings.NewReader(out.String()), nil
}
