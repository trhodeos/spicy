package spicy

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

type OutputFileRunner struct {
	runner             Runner
	expectedOutputFile string
}

func NewOutputFileRunner(r Runner, outputFile string) OutputFileRunner {
	return OutputFileRunner{runner: r, expectedOutputFile: outputFile}
}

func (e OutputFileRunner) Run(r io.Reader, args []string) (io.Reader, error) {
	_, err := e.runner.Run(r, args)
	if err != nil {
		return nil, err
	}
	return os.Open(e.expectedOutputFile)
}

type MappedFileRunner struct {
	runner        Runner
	inputFileArgs map[string]io.Reader
	outputFileArg string
}

func NewMappedFileRunner(r Runner, inputFileArgs map[string]io.Reader, outputFileArg string) MappedFileRunner {
	return MappedFileRunner{runner: r, inputFileArgs: inputFileArgs, outputFileArg: outputFileArg}
}

func writeTempFile(r io.Reader, prefix string) (string, error) {
	tmpfile, err := ioutil.TempFile("", prefix)
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(tmpfile.Name())
	if err != nil {
		return "", err
	}
	log.Debugln("Writing file for prefix %s to %s", prefix, path)
	_, err = io.Copy(tmpfile, r)
	if err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}
	return path, nil
}

func (e MappedFileRunner) Run(r io.Reader, args []string) (io.Reader, error) {
	var newArgs []string = make([]string, len(args))
	for i, arg := range args {
		if _, ok := e.inputFileArgs[arg]; ok {
			tempFile, err := writeTempFile(e.inputFileArgs[arg], arg)
			if err != nil {
				return nil, err
			}
			newArgs[i] = tempFile
		} else {
			newArgs[i] = args[i]
		}
	}
	_, err := e.runner.Run(r, args)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(e.outputFileArg)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}
