package main
import (
  "flag"
  "bufio"
  "os"
  "github.com/trhodeos/spicy"
)
const (
  defines_text = "Defines passed to cpp."
  includes_text = "Includes passed to cpp.."
  undefine_text = "Includes passed to cpp.."
  verbose_text = "If true, be verbose."
  verbose_link_editor_text = "If true, be verbose when link editing."
  disable_overlapping_section_check_text = ""
  romsize_text = "Rom size (MBits)"
  filldata_text = "filldata byte"
  bootstrap_filename_text = "Bootstrap file"
  header_filename_text = "Header file"
  pif_bootstrap_filename_text = "Pif bootstrap file"
  rom_image_file_text = "Rom image filename"
  spec_file_text = "Spec file to use for making the image"
  ld_command_text = "ld command to use"
  cpp_command_text = "cpp command to use"
)

var (
  defines = flag.String("D", "", defines_text)
  includes = flag.String("I", "", includes_text)
  undefine = flag.String("U", "", undefine_text)
  verbose = flag.Bool("d", false, verbose_text)
  link_editor_verbose = flag.Bool("m", false, verbose_link_editor_text)
  disable_overlapping_section_check = flag.Bool("o", false, disable_overlapping_section_check_text)
  romsize_mbits = flag.Int("s", -1, romsize_text)
  filldata = flag.Int("f", 0x0, filldata_text)
  bootstrap_filename = flag.String("b", "Boot", bootstrap_filename_text)
  header_filename = flag.String("h", "romheader", header_filename_text)
  pif_bootstrap_filename = flag.String("p", "pif2Boot", pif_bootstrap_filename_text)
  rom_image_file = flag.String("r", "rom", rom_image_file_text)

  // Non-standard options. Should all be optional.
  ld_command = flag.String("ld_command", "mips-elf-ld", ld_command_text)
  cpp_command = flag.String("cpp_command", "mips-elf-cpp", cpp_command_text)
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
  flag.Parse()

  f, err := os.Open(flag.Arg(0))
  if err != nil {
    panic(err)
  }

  spec, err := spicy.ParseSpec(bufio.NewReader(f))
  if err != nil { panic(err) }
  err = spec.CheckValidity()
  if err != nil { panic(err) }
  err = spicy.LinkSpec(spec, *ld_command)
  if err != nil { panic(err) }
}
