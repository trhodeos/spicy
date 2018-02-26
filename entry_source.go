package spicy

import (
	"bytes"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

func createEntrySource(bootSegment *Segment) (string, error) {
	t := `
	.text
	.globl	_start
_start:
	la	$8,_{{.Name}}SegmentBssStart
	la	$9,_{{.Name}}SegmentBssSize
1:
	sw	$0, 0($8)
	sw	$0, 4($8)
	addi	$8, 8
	addi	$9, 0xfff8
	bne	$9, $0, 1b
	la	$10, {{.Entry}} + 0
	la	$29,{{.StackInfo.Start}} + {{.StackInfo.Offset}}
	jr	$10

`
	tmpl, err := template.New("test").Parse(t)
	if err != nil {
		return "", err
	}
	b := &bytes.Buffer{}
	err = tmpl.Execute(b, bootSegment)
	return b.String(), err
}

func generateEntryScript(w *Wave) (string, error) {
	glog.V(1).Infoln("Starting to generate entry script.")
	content, err := createEntrySource(w.GetBootSegment())
	if err != nil {
		return "", err
	}
	glog.V(2).Infoln("Entry script generated:\n", content)
	tmpfile, err := ioutil.TempFile("", "entry-script")
	path, err := filepath.Abs(tmpfile.Name())
	if err != nil {
		return "", err
	}
	glog.V(1).Infoln("Writing script to", path)
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}
	glog.V(1).Infoln("Script written.")
	return path, nil
}

func CreateEntryBinary(w *Wave, as_command string, ld_command string) (*os.File, error) {
	name := w.Name
	glog.Infof("Creating entry for \"%s\".", name)
	entry_source_path, err := generateEntryScript(w)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(as_command, "-non_shared", entry_source_path)
	var out bytes.Buffer
	var errout bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errout
	err = cmd.Run()
	if glog.V(2) {
		glog.V(2).Info("as stdout: ", out.String())
	}
	if err != nil {
		glog.Error("Error running as. Stderr output: ", errout.String())
	}
	return nil, err
}
