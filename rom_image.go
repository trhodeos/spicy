package spicy

import (
        "errors"
        "fmt"
	"os"
)

func fill(f *os.File, fill byte, start int, end int) error {
  if start > end {
    return errors.New(fmt.Sprintf("When filling bytes, start [%d] is greater than end [%d].", start, end))
  }
  fill_bytes := make([]byte, end - start)
  for i := 0; i < len(fill_bytes); i++ {
    fill_bytes[i] = fill
  }
  _, err := f.WriteAt(fill_bytes, int64(start))
  return err
}

func WriteRomImage(
  romname string,
  fillByte byte,
  romheader []byte,
  pif2boot []byte,
  boot []byte,
  font []byte,
  entry []byte,
  code []byte,
  raw []byte) error {
    f, err := os.Create(romname)
    if err != nil {
      return err
    }
    defer f.Close()
    if len(romheader) > 0x40 {
      return errors.New(
        fmt.Sprintf("Romheader is of size 0x%x, cannot be larger than 0x40", len(romheader)))
    }
    f.WriteAt(romheader, 0)
    err = fill(f, fillByte, len(romheader), 0x40)
    if err != nil {
      return err
    }

    f.WriteAt(pif2boot, 0x40)
    err = fill(f, fillByte, 0x40 + len(pif2boot), 0xFFF)
    if err != nil {
      return err
    }

    f.WriteAt(pif2boot, 0x40)
    err = fill(f, fillByte, 0x40 + len(pif2boot), 0xFFF)
    if err != nil {
      return err
    }

	return nil
}
