package spicy

import (
	"bytes"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"text/template"
)

var ldArgs = []string{"-G 0", "-S", "-noinhibit-exec", "-nostartfiles", "-nodefaultlibs", "-nostdinc", "-M"}

func createLdScript(w *Wave) (io.Reader, error) {
	t := `
ENTRY(_start)
MEMORY {
    {{range .ObjectSegments}}
    {{.Name}}.RAM (RX) : ORIGIN = {{.Positioning.Address}}, LENGTH = 0x400000
    {{.Name}}.bss.RAM (RW) : ORIGIN = {{.Positioning.Address}}, LENGTH = 0x400000
    {{end}}
}
SECTIONS {
    ..generatedStartEntry 0x80000400 : AT(0x1000)
    {
      a.out (.text)
    }

    _RomSize = 0x1050;
    _RomStart = _RomSize;
    {{range $index, $seg := .ObjectSegments -}}
    _{{$seg.Name}}SegmentRomStart = _RomSize;
    ..{{$seg.Name}} {{$seg.Positioning.Address}} : AT(_RomSize)
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
  /DISCARD/ :
  {
        *(.MIPS.abiflags*)
  }
  _RomEnd = _RomSize;
}
`
	tmpl, err := template.New("test").Parse(t)
	if err != nil {
		return nil, err
	}
	b := &bytes.Buffer{}
	err = tmpl.Execute(b, w)
	if err == nil {
		log.Debugln("Ld script generated:\n", b.String())
	}
	return b, err
}

func LinkSpec(w *Wave, ld Runner, entry io.Reader) (io.Reader, error) {
	name := w.Name
	log.Infof("Linking spec \"%s\".", name)
	ldscript, err := createLdScript(w)
	if err != nil {
		return nil, err
	}
	outputPath := fmt.Sprintf("%s.out", name)
	mappedInputs := map[string]io.Reader{
		"ld-script": ldscript,
	}
	return NewMappedFileRunner(ld, mappedInputs, outputPath).Run( /* stdin=*/ nil, append(ldArgs, "-dT", "ld-script", "-o", outputPath))
}
func TempFileName(suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), hex.EncodeToString(randBytes)+suffix)
}

func BinarizeObject(obj io.Reader, objcopy Runner) (io.Reader, error) {
	outputBin := TempFileName(".bin")
	mappedInputs := map[string]io.Reader{
		"objFile": obj,
	}
	return NewMappedFileRunner(objcopy, mappedInputs, outputBin).Run( /* stdin=*/ nil, []string{"-O", "binary", "objFile", outputBin})
}
