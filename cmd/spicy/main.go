package main

import (
	flag "github.com/ogier/pflag"
	log "github.com/sirupsen/logrus"
	"github.com/trhodeos/n64rom"
	"github.com/trhodeos/spicy"
	"io/ioutil"
	"os"
)

const (
	defines_text                           = "Defines passed to cpp."
	includes_text                          = "Includes passed to cpp."
	undefine_text                          = "Undefines passed to cpp.."
	verbose_text                           = "If true, be verbose."
	verbose_link_editor_text               = "If true, be verbose when link editing."
	disable_overlapping_section_check_text = "If true, disable overlapping section checks."
	romsize_text                           = "Rom size (MBits)"
	filldata_text                          = "filldata byte"
	bootstrap_filename_text                = "Bootstrap file (not currently used)"
	header_filename_text                   = "Header file (not currently used)"
	pif_bootstrap_filename_text            = "Pif bootstrap file (not currently used)"
	rom_image_file_text                    = "Rom image filename"
	spec_file_text                         = "Spec file to use for making the image"
	ld_command_text                        = "ld command to use"
	as_command_text                        = "as command to use"
	cpp_command_text                       = "cpp command to use"
	objcopy_command_text                   = "objcopy command to use"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "List of strings"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var defineFlags arrayFlags
var includeFlags arrayFlags
var undefineFlags arrayFlags

var (
	verbose                           = flag.BoolP("verbose", "d", false, verbose_text)
	link_editor_verbose               = flag.BoolP("verbose_linking", "m", false, verbose_link_editor_text)
	disable_overlapping_section_check = flag.BoolP("disable_overlapping_section_checks", "o", false, disable_overlapping_section_check_text)
	romsize_mbits                     = flag.IntP("romsize", "s", -1, romsize_text)
	filldata                          = flag.IntP("filldata_byte", "f", 0x0, filldata_text)
	bootstrap_filename                = flag.StringP("bootstrap_file", "b", "Boot", bootstrap_filename_text)
	header_filename                   = flag.StringP("romheader_file", "h", "romheader", header_filename_text)
	pif_bootstrap_filename            = flag.StringP("pif2boot_file", "p", "pif2Boot", pif_bootstrap_filename_text)
	rom_image_file                    = flag.StringP("rom_name", "r", "rom.n64", rom_image_file_text)
	elf_file                          = flag.StringP("rom_elf_name", "e", "rom.out", rom_image_file_text)

	// Non-standard options. Should all be optional.
	ld_command      = flag.String("ld_command", "mips64-elf-ld", ld_command_text)
	as_command      = flag.String("as_command", "mips64-elf-as", as_command_text)
	cpp_command     = flag.String("cpp_command", "mips64-elf-gcc", cpp_command_text)
	objcopy_command = flag.String("objcopy_command", "mips64-elf-objcopy", objcopy_command_text)
	font_filename   = flag.String("font_filename", "font", "Font filename")
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
	flag.VarP(&defineFlags, "define", "D", defines_text)
	flag.VarP(&includeFlags, "include", "I", includes_text)
	flag.VarP(&undefineFlags, "undefine", "U", undefine_text)
	flag.Parse()
	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gcc := spicy.NewRunner(*cpp_command)
	ld := spicy.NewRunner(*ld_command)
	as := spicy.NewRunner(*as_command)
	objcopy := spicy.NewRunner(*objcopy_command)
	preprocessed, err := spicy.PreprocessSpec(f, gcc, includeFlags, defineFlags, undefineFlags)
	spec, err := spicy.ParseSpec(preprocessed)
	if err != nil {
		panic(err)
	}

	rom, err := n64rom.NewBlankRomFile(byte(*filldata))
	if err != nil {
		panic(err)
	}
	for _, w := range spec.Waves {
		for _, seg := range w.RawSegments {
			for _, include := range seg.Includes {
				f, err := os.Open(include)
				if err != nil {
					panic(err)
				}
				spicy.CreateRawObjectWrapper(f, include+".o", ld)
			}
		}
		entry, err := spicy.CreateEntryBinary(w, as)
		if err != nil {
			panic(err)
		}
		linked_object, err := spicy.LinkSpec(w, ld, entry)
		if err != nil {
			panic(err)
		}
		binarized_object, err := spicy.BinarizeObject(linked_object, objcopy)
		if err != nil {
			panic(err)
		}

		binarized_object_bytes, err := ioutil.ReadAll(binarized_object)
		if err != nil {
			panic(err)
		}
		rom.WriteAt(binarized_object_bytes, n64rom.CodeStart)
		if err != nil {
			panic(err)
		}
	}
	out, err := os.Create(*rom_image_file)
	if err != nil {
		panic(err)
	}
	// Pad the rom if necessary.
	if *romsize_mbits > 0 {
		minSize := int64(1000000 * *romsize_mbits / 8)
		_, err := out.WriteAt([]byte{0}, minSize)
		if err != nil {
			panic(err)
		}
	}
	_, err = rom.Save(out)
	if err != nil {
		panic(err)
	}
	err = out.Close()
	if err != nil {
		panic(err)
	}
}
