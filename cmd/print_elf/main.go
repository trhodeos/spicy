package main

import (
	"debug/elf"
	"fmt"
	"os"
)

func main() {
	path := os.Args[1]
	f, err := elf.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Printf("%+v\n", f)
	for _, sec := range f.Sections {
		fmt.Printf("%+v\n", sec.SectionHeader)
	}
}
