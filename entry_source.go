package spicy

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

var compileArgs = []string{"-march=vr4300", "-mtune=vr4300", "-mgp32", "-mfp32", "-non_shared"}

func createEntrySource(bootSegment *Segment) (string, error) {
	t := `
	.text
	.global	_start
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
	log.Debug("Starting to generate entry script.")
	content, err := createEntrySource(w.GetBootSegment())
	if err != nil {
		return "", err
	}
	log.Debugf("Entry script generated:\n%s", content)
	tmpfile, err := ioutil.TempFile("", "entry-script")
	path, err := filepath.Abs(tmpfile.Name())
	if err != nil {
		return "", err
	}
	log.Debug("Writing script to", path)
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}
	log.Debug("Script written.")
	return path, nil
}

func CreateEntryBinary(w *Wave, as Runner) (*os.File, error) {
	name := w.Name
	log.Infof("Creating entry for \"%s\".", name)
	entrySourcePath, err := generateEntryScript(w)
	if err != nil {
		return nil, err
	}
	_, err = as.Run( /* stdin=*/ nil, append(compileArgs, entrySourcePath))
	if err != nil {
		return nil, err
	}
	// TODO(trhodeos): Make this not nil.
	return nil, err
}
