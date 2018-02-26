package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
        "io/ioutil"
	"github.com/trhodeos/spicy"
)

func main() {
  path := os.Args[1]

  file, err := os.Open(path)
  if err != nil {
    log.Fatal("Error while opening file", err)
  }

  defer file.Close()

  fmt.Printf("%s opened\n", path)

  b, err := ioutil.ReadAll(file)
  if err != nil {
    log.Fatal("Error while reading file", err)
  }
  buffer := bytes.NewBuffer(b)

  fileheader := spicy.FileHeader{}
  err = binary.Read(buffer, binary.BigEndian, &fileheader)
  if err != nil {
    log.Fatal("binary.Read failed", err)
  }
  fmt.Printf("Parsed file header:\n%+v\n", fileheader)

  objheader := spicy.AoutHeader{}
  err = binary.Read(buffer, binary.BigEndian, &objheader)
  if err != nil {
    log.Fatal("binary.Read failed", err)
  }
  fmt.Printf("Parsed obj header:\n%+v\n", objheader)

  var i uint16
  for i = 0; i < fileheader.NumSections; i++ {
    secheader := spicy.SectionHeader{}
    err = binary.Read(buffer, binary.BigEndian, &secheader)
    if err != nil {
      log.Fatal("binary.Read failed", err)
    }
    fmt.Printf("Parsed sec header %d:\n%+v\n", i, secheader)
  }
}
