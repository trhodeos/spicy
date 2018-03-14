package spicy

import (
	"bytes"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
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

func CreateEntryBinary(w *Wave, as_command string, ld_command string, objcopy_command string, linked_obj string) (*os.File, error) {
	name := w.Name
	glog.Infof("Creating entry for \"%s\".", name)
	entry_source_path, err := generateEntryScript(w)
	if err != nil {
		return nil, err
	}
	err = RunCmd(as_command, "-mgp32", "-mfp32", "-march=vr4300", "-non_shared", entry_source_path)
	if err != nil {
		return nil, err
	}
	tmpfile, err := ioutil.TempFile("", "linked-entry-script")
	path, err := filepath.Abs(tmpfile.Name())
	if err != nil {
		tmpfile.Close()
		return nil, err
	}
	err = RunCmd(ld_command, "-R", linked_obj, "-o", path, "a.out")
	if err != nil {
		tmpfile.Close()
		return nil, err
	}
	defer tmpfile.Close()
	outfile, err := ioutil.TempFile("", "binarized-entry-script")
	outpath, err := filepath.Abs(outfile.Name())
	err = RunCmd(objcopy_command, "-O", "binary", path, outpath)
	return outfile, err
}
