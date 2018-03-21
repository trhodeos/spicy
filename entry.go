package spicy

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"io"
	"text/template"
)

var compileArgs = []string{"-march=vr4300", "-mtune=vr4300", "-mgp32", "-mfp32", "-non_shared"}

func createEntrySource(bootSegment *Segment) (io.Reader, error) {
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
		return nil, err
	}
	b := &bytes.Buffer{}
	err = tmpl.Execute(b, bootSegment)
	return b, err
}

func CreateEntryBinary(w *Wave, as Runner) (io.Reader, error) {
	name := w.Name
	log.Infof("Creating entry for \"%s\".", name)
	entrySource, err := createEntrySource(w.GetBootSegment())
	if err != nil {
		return nil, err
	}
	return NewOutputFileRunner(as, "a.out").Run(entrySource, append(compileArgs, "-"))
}
