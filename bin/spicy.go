package main
import (
  "encoding/json"
  "os"
  "github.com/alecthomas/kingpin"
  "github.com/trhodeos/spicy"
)

var (
  defines = kingpin.Flag("define", "Defines passed to cpp.").Short('D').String()
  includes = kingpin.Flag("include", "Includes passed to cpp..").Short('I').String()
  undefine = kingpin.Flag("undefine", "Includes passed to cpp..").Short('U').String()
  verbose = kingpin.Flag("verbose", "If true, be verbose.").Short('d').Bool()
  link_editor_verbose = kingpin.Flag("link_editor_verbose", "If true, be verbose when link editing.").Short('m').Bool()
  disable_overlapping_section_check = kingpin.Flag("disable_overlapping_section_check", "").Short('o').Bool()
  romsize_mbits = kingpin.Flag("romsize", "Rom size (MBits)").Short('s').Int()
  filldata = kingpin.Flag("filldata", "filldata").Short('f').Int()
  bootstrap_filename = kingpin.Flag("bootstrap_filename", "Bootstrap file").Short('b').Default("Boot").String()
  header_filename = kingpin.Flag("header_filename", "Header filename").Short('h').Default("romheader").String()
  pif_bootstrap_filename = kingpin.Flag("pif_bootstrap_filename", "Pif bootstrap filename").Short('p').Default("pif2Boot").String()
  rom_image_file = kingpin.Flag("rom_image_filename", "Rom image filename").Short('r').Default("rom").String()
  spec_file = kingpin.Arg("spec_file", "Spec file to use for making the image").Required().String()
)
/*
-Dname[=def] Is passed to cpp(1) for use during its invocation.
-Idirectory Is passed to cpp(1) for use during its invocation.
Uname Is passed to cpp(1) for use during its invocation.
-d Gives a verbose account of all the actions that makerom takes, leaving temporary files created that are ordinarily deleted.
-m Prints a link editor map to standard output for diagnostic purposes.
-o Disables checking of overlapping sections. By default, segments with direct-mapped CPU addresses are checked to ensure that the underlying physical memory mappings do not conflict.
-b <bootstrap filename> Overrides the default filename (/usr/lib/PR/Boot). This file must be in COFF format, and is loaded as 1K bytes into the ramrom memory.
-h <header filename> Overrides the default ROM header filename (/usr/lib/PR/romheader). This file is in ASCII format, and each character is converted to hex and loaded in sequence, starting at the beginning of ramrom memory. Currently only 32 bytes are used.
-s <romsize (Mbits)> Creates a ROM image with the specified size. This option is required for making the real Game Pak.
-f <filldata (0x0 - 0xff)> Sets the fill pattern for "holes" within a ROM image. The argument filldata is a one-byte hexadecimal constant. Use this option when you create a ROM image using the -s option. It is required for making the real Game Pak.
-p <pif bootstrap file> Overrides the pif bootstrap filename (/usr/lib/PR/pif2Boot). This file must be in COFF format. It is loaded as 4K bytes into the ramrom memory.
-r Provides an alternate ROM image file; the default is 'rom'.
-B 0 An option that concerns only games supported by 64DD. Using this option creates a startup game. For information on startup games, please see Section 15.1, "Restarting," in the N64 Disk Drive Programming Manual.
*/

func main() {
  kingpin.Parse()
  var spec, err = spicy.ParseSpec(os.Stdin)
  if err != nil { panic(err) }

  json.NewEncoder(os.Stdout).Encode(spec)
}
