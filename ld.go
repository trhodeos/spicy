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
    ram (RX) : ORIGIN = 0xFFFFFFFF80000400, LENGTH = 0x77FFFBFF
}
SECTIONS {
    ..generatedStartEntry 0xFFFFFFFF80000400 :
    {
      a.out (.text)
    } > ram
    _RomSize = 0x1050;
    _RomStart = _RomSize;
    {{range .ObjectSegments -}}
    _{{.Name}}SegmentRomStart = _RomSize;
    ..{{.Name}}
    {{if ne .Positioning.AfterSegment ""}}
        ADDR(..{{.Positioning.AfterSegment}}.bss) + SIZEOF(..{{.Positioning.AfterSegment}}.bss)
    {{else if ne (index .Positioning.AfterMinSegment 0) ""}}
        MIN(
	  ADDR(..{{index .Positioning.AfterMinSegment 0}}.bss) + SIZEOF(..{{index .Positioning.AfterMinSegment 0}}.bss),
          ADDR(..{{index .Positioning.AfterMinSegment 1}}.bss) + SIZEOF(..{{index .Positioning.AfterMinSegment 1}}.bss))
    {{else if ne (index .Positioning.AfterMaxSegment 0) ""}}
        MAX(
	  ADDR(..{{index .Positioning.AfterMaxSegment 0}}.bss) + SIZEOF(..{{index .Positioning.AfterMaxSegment 0}}.bss),
          ADDR(..{{index .Positioning.AfterMaxSegment 1}}.bss) + SIZEOF(..{{index .Positioning.AfterMaxSegment 1}}.bss))
    {{else if not (eq .Positioning.Address 0)}}
      {{.Positioning.Address}}
    {{end}}
    :
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
        {{.}} (.rodata*)
      {{end}}
      {{range .Includes -}}
        {{.}} (.sdata)
      {{end}}
      . = ALIGN(0x10);
      _{{.Name}}SegmentDataEnd = .;
    } > ram
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
    } > ram
    _RomSize += ( _{{.Name}}SegmentBssEnd - _{{.Name}}SegmentBssStart );
    _{{.Name}}SegmentBssSize = ( _{{.Name}}SegmentBssEnd - _{{.Name}}SegmentBssStart );
  {{ end }}
  {{range .RawSegments -}}
    _{{.Name}}SegmentRomStart = _RomSize;
    ..{{.Name}} :
    {
      _{{.Name}}SegmentDataStart = .;
      {{range .Includes -}}
      "{{.}}.o"
      {{end}}
      _{{.Name}}SegmentDataEnd = .;
    } > ram
    _RomSize += ( _{{.Name}}SegmentDataEnd - _{{.Name}}SegmentDataStart );
    _{{.Name}}SegmentRomEnd = _RomSize;
  {{ end }}
  /DISCARD/ :
  {
    *(.MIPS.abiflags*)
    *(.mdebug*)
    *(.gnu.attributes*)
    *(.pdr*)
    *(.reginfo*)
    *(.comment*)
    *(.options*)
    *(.gptab*)
    *(.note*)
    *(.rel.dyn*)
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

func CreateRawObjectWrapper(r io.Reader, outputName string, ld Runner) (io.Reader, error) {
	mappedInputs := map[string]io.Reader{
		"input": r,
	}
	return NewMappedFileRunner(ld, mappedInputs, outputName).Run( /* stdin=*/ nil, []string{"-r", "-b", "binary", "-o", outputName, "input"})
}
