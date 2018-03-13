package spicy

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"os/exec"
	"strings"
)

func RunCmd(command string, args ...string) error {
	fmt.Printf("About to run %s %s\n", command, strings.Join(args, " "))
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	var errout bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errout
	err := cmd.Run()
	if glog.V(2) {
		glog.V(2).Info(command, " stdout: ", out.String())
	}
	if err != nil {
		glog.Error("Error running ", command, ". Stderr output: ", errout.String())
	}
	return err
}
