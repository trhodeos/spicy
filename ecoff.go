package spicy

type FileHeader struct {
  Magic uint16
  NumberSections uint16
  TimeDate int32
  SymbolsPointer int32
  NumberSymbols int32
  OptionalHeader uint16
  Flags uint16
}

type AoutHeader struct {
  Magic int16
  Vstamp int16
  TextSize int32
  DataSize int32
  BssSize int32
  Entry int32
  TextStart int32
  DataStart int32
  BssStart int32
  GprMask int32
  CprMask [4]int32
  GpValue int32
}

type SectionHeader struct {
  Name [8]uint8
  PhysicalAddress int32
  VirtualAddress int32
  Size int32
  SectionPointer int32
  RelocationsPointer int32
  LnnoPtr int32
  NumRelocations uint16
  NumLnno uint16
  Flags int32
}
