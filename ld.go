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
    ..generatedStartEntry 0xFFFFFFFF80000400 : AT(0xFFFFFFFF80000400)
    {
      a.out (.text)
    }

    _RomSize = 0x1050;
    _RomStart = _RomSize;
    {{range .ObjectSegments -}}
    _{{.Name}}SegmentRomStart = _RomSize;
    ..{{.Name}} {{.Positioning.Address}} {{if .Positioning.NoLoad }} (NOLOAD) {{end}} :
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
    } AT> ram
    _RomSize += ( _{{.Name}}SegmentBssEnd - _{{.Name}}SegmentBssStart );
    _{{.Name}}SegmentBssSize = ( _{{.Name}}SegmentBssEnd - _{{.Name}}SegmentBssStart );
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
