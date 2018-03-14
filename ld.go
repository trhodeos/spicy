package spicy

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

func createLdScript(w *Wave) (string, error) {
	t := `
MEMORY {
    {{range .ObjectSegments}}
    {{.Name}}.RAM (RX) : ORIGIN = {{.Positioning.Address}}, LENGTH = 0x400000
    {{.Name}}.bss.RAM (RW) : ORIGIN = {{.Positioning.Address}}, LENGTH = 0x400000
    {{end}}
}
SECTIONS {
    _RomSize = 0x1050;
    _RomStart = _RomSize;
    {{range $index, $seg := .ObjectSegments -}}
    _{{$seg.Name}}SegmentRomStart = _RomSize;
    ..{{$seg.Name}} {{if eq $index 0 }}{{$seg.Positioning.Address}}{{ end }}: AT( _RomSize )
    {
        _{{$seg.Name}}SegmentStart = .;
        . = ALIGN(0x10);
        _{{$seg.Name}}SegmentTextStart = .;
            {{range $seg.Includes -}}
              {{.}} (.text)
            {{end}}
        _{{$seg.Name}}SegmentTextEnd = .;
        _{{$seg.Name}}SegmentDataStart = .;
            {{range $seg.Includes -}}
              {{.}} (.data)
            {{end}}
            {{range $seg.Includes -}}
              {{.}} (.rodata)
            {{end}}
            {{range $seg.Includes -}}
              {{.}} (.sdata)
            {{end}}
        . = ALIGN(0x10);
        _{{$seg.Name}}SegmentDataEnd = .;
    } > {{$seg.Name}}.RAM
    _RomSize += ( _{{$seg.Name}}SegmentDataEnd - _{{$seg.Name}}SegmentTextStart );
    _{{$seg.Name}}SegmentRomEnd = _RomSize;

    ..{{$seg.Name}}.bss ADDR(..{{$seg.Name}}) + SIZEOF(..{{$seg.Name}}) (NOLOAD) :
    {
        . = ALIGN(0x10);
        _{{$seg.Name}}SegmentBssStart = .;
            {{range $seg.Includes -}}
              {{.}} (.sbss)
            {{end}}
            {{range $seg.Includes -}}
              {{.}} (.scommon)
            {{end}}
            {{range $seg.Includes -}}
              {{.}} (.bss)
            {{end}}
            {{range $seg.Includes -}}
              {{.}} (COMMON)
            {{end}}
        . = ALIGN(0x10);
        _{{$seg.Name}}SegmentBssEnd = .;
        _{{$seg.Name}}SegmentEnd = .;
    } > {{$seg.Name}}.bss.RAM
    _{{$seg.Name}}SegmentBssSize = ( _{{$seg.Name}}SegmentBssEnd - _{{$seg.Name}}SegmentBssStart );
  {{ end }}
  {{range $index, $seg := .RawSegments -}}
    _{{$seg.Name}}SegmentRomStart = _RomSize;
    ..{{$seg.Name}} : AT( _RomSize )
    {
        _{{$seg.Name}}SegmentStart = .;
        . = ALIGN(0x10);
        _{{$seg.Name}}SegmentTextStart = .;
            {{range $seg.Includes -}}
              {{.}} (.text)
            {{end}}
        _{{$seg.Name}}SegmentTextEnd = .;
        _{{$seg.Name}}SegmentDataStart = .;
            {{range $seg.Includes -}}
              {{.}} (.data)
            {{end}}
            {{range $seg.Includes -}}
              {{.}} (.rodata)
            {{end}}
            {{range $seg.Includes -}}
              {{.}} (.sdata)
            {{end}}
        . = ALIGN(0x10);
        _{{$seg.Name}}SegmentDataEnd = .;
    } > {{$seg.Name}}.RAM
    _RomSize += ( _{{$seg.Name}}SegmentDataEnd - _{{$seg.Name}}SegmentTextStart );
    _{{$seg.Name}}SegmentRomEnd = _RomSize;

    ..{{$seg.Name}}.bss ADDR(..{{$seg.Name}}) + SIZEOF(..{{$seg.Name}}) (NOLOAD) : AT ( _RomSize )
    {
        . = ALIGN(0x10);
        _{{$seg.Name}}SegmentBssStart = .;
            {{range $seg.Includes -}}
              {{.}} (.sbss)
            {{end}}
            {{range $seg.Includes -}}
              {{.}} (.scommon)
            {{end}}
            {{range $seg.Includes -}}
              {{.}} (.bss)
            {{end}}
            {{range $seg.Includes -}}
              {{.}} (COMMON)
            {{end}}
        . = ALIGN(0x10);
        _{{$seg.Name}}SegmentBssEnd = .;
        _{{$seg.Name}}SegmentEnd = .;
    } > {{$seg.Name}}.bss.RAM
    _{{$seg.Name}}SegmentBssSize = ( _{{$seg.Name}}SegmentBssEnd - _{{$seg.Name}}SegmentBssStart );
  {{ end }}
  /DISCARD/ :
  {
        *(.MIPS.abiflags*)
	  }
  _RomEnd = _RomSize;
}
`
	tmpl, err := template.New("test").Parse(t)
	if err != nil {
		return "", err
	}
	b := &bytes.Buffer{}
	err = tmpl.Execute(b, w)
	return b.String(), err
}

func generateLdScript(w *Wave) (string, error) {
	glog.V(1).Infoln("Starting to generate ld script.")
	content, err := createLdScript(w)
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

func LinkSpec(w *Wave, ld_command string) (string, error) {
	name := w.Name
	glog.Infof("Linking spec \"%s\".", name)
	ld_path, err := generateLdScript(w)
	if err != nil {
		return "", err
	}
	output_path := fmt.Sprintf("%s.out", name)
	err = RunCmd(ld_command, "-S", "-nostartfiles", "-nodefaultlibs", "-nostdinc", "-dT", ld_path, "-o", output_path, "-M")
	if err != nil {
		return "", err
	}
	return output_path, err
}

func BinarizeObject(obj_path string, objcopy_command string) (*os.File, error) {
	output_bin := fmt.Sprintf("%s.bin", obj_path)
	err := RunCmd(objcopy_command, "-O", "binary", obj_path, output_bin)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(output_bin)
	if err != nil {
		return nil, err
	}
	_, err = file.Seek(0x1050, os.SEEK_CUR)
	return file, err
}
