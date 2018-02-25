package spicy

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"text/template"
        "bytes"
)

func createLdScript(s *Spec) (string, error) {
	t := `
MEMORY {
    {{range .Segments}}
    {{.Name}}.RAM (RX) : ORIGIN = 0x80000450, LENGTH = 0x400000
    {{.Name}}.bss.RAM (RW) : ORIGIN = 0x80000450, LENGTH = 0x400000
    {{end}}
}
SECTIONS {
    _RomSize = 0x1050;
    _RomStart = _RomSize;
  {{range .Segments -}}
    _{{.Name}}SegmentRomStart = _RomSize;
    ..{{.Name}} 0x80000450 :
    {
        _{{.Name}}SegmentStart = .;
        . = ALIGN(0x10);
        _{{.Name}}SegmentTextStart = .;
            {{range .Includes -}}
            {{.}} (.text)
            {{end}}
        _{{.Name}}SegmentTextEnd = .;
        _{{.Name}}SegmentDataStart = .;
            {{range .Includes -}}
            {{.}} (.data)
            {{end}}
            {{range .Includes -}}
            {{.}} (.rodata)
            {{end}}
            {{range .Includes -}}
            {{.}} (.sdata)
            {{end}}
        . = ALIGN(0x10);
        _{{.Name}}SegmentDataEnd = .;
    } > {{.Name}}.RAM
    _RomSize += ( _{{.Name}}SegmentDataEnd - _{{.Name}}SegmentTextStart );
    _{{.Name}}SegmentRomEnd = _RomSize;

    ..{{.Name}}.bss ADDR(..{{.Name}}) + SIZEOF(..{{.Name}}) (NOLOAD) :
    {
        . = ALIGN(0x10);
        _{{.Name}}SegmentBssStart = .;
            {{range .Includes -}}
            {{.}} (.sbss)
            {{end}}
            {{range .Includes -}}
            {{.}} (.scommon)
            {{end}}
            {{range .Includes -}}
            {{.}} (.bss)
            {{end}}
            {{range .Includes -}}
            {{.}} (COMMON)
            {{end}}
        . = ALIGN(0x10);
        _{{.Name}}SegmentBssEnd = .;
        _{{.Name}}SegmentEnd = .;
    } > {{.Name}}.bss.RAM
    _{{.Name}}SegmentBssSize = ( _{{.Name}}SegmentBssEnd - _{{.Name}}SegmentBssStart );
  {{- end}}
}
`
	tmpl, err := template.New("test").Parse(t)
	if err != nil {
		return "", err
	}
        b := &bytes.Buffer{}
	err = tmpl.Execute(b, s)
	return b.String(), err
}

func generateLdScript(s *Spec) (string, error) {
	glog.V(1).Infoln("Starting to generate ld script.")
	content, err := createLdScript(s)
	if err != nil {
		return "", err
	}
	glog.V(2).Infoln("Ld script generated:\n", content)
	tmpfile, err := ioutil.TempFile("", "ld-script")
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

func LinkSpec(s *Spec, ld_command string) error {
	name := s.Waves[0].Name
	glog.Infof("Linking spec \"%s\".", name)
	ld_path, err := generateLdScript(s)
	if err != nil {
		return err
	}
	cmd := exec.Command(ld_command, "-G 0", "-noinhibit-exec", "-T", ld_path, "-o", fmt.Sprintf("%s.out", name), "-M")
        var out bytes.Buffer
        var errout bytes.Buffer
        cmd.Stdout = &out
        cmd.Stderr = &errout
	err = cmd.Run()
        if glog.V(2) {
          glog.V(2).Info("Ld stdout: ", out.String())
        }
        if err != nil {
          glog.Error(errout.String())
        }
        return err
}

