package main

import (
  "encoding/json"

  "os"

  "github.com/alecthomas/participle"
)

type Value struct {
  Symbol string `@Ident |`
  Constant int `@Int`
}

type StackValue struct {
  Value Value `@@`
  AddedValues []*Value `{"+" @@}`
}

type SegmentStatement struct {
/*
 :name <segmentName>
                |address <constant>
                |after <segmentName>
                |after max[<segmentName>,<segmentName>]
                |after min[<segmentName>,<segmentName>]
                |include <filename>
                |maxsize <constant>
                |align <constant>
                |flags <flagList>
                |number <constant>
                |entry <symbol>
                |stack <stackValue>
*/
  Name string      `"name" @String |`
  Address     int `"address" @Int |`
  After       string      `"after" @String |`
  Include     string     `"include" @String |`
  MaxSize     string     `"maxsize" @String |`
  Align       string     `"align" @String |`
  Flags       []*string     `"flags" { "BOOT" | "OBJECT" | "RAW" } |`
  Number      int `"number" @Int|`
  Entry string     `"entry" @Ident |`
  Stack StackValue `"stack" @@`
}

type Segment struct {
  SegmentStatementList []*SegmentStatement `"beginseg" { @@ } "endseg"`
}

type WaveStatement struct {
/*
 :name <waveName>
                |include <segmentName>
*/
  Name string      `"name" @String |`
  Include     string     `"include" @String`
}

type Wave struct {
  WaveStatementList []*WaveStatement `"beginwave" { @@ } "endwave"`
}

type Spec struct {
  SegmentList []*Segment `{ @@ }`
  WaveList []*Wave `{ @@ }`
}

func main() {
  parser, err := participle.Build(&Spec{}, nil)
  if err != nil { panic(err) }

  spec := &Spec{}
  err = parser.Parse(os.Stdin, spec)
  if err != nil { panic(err) }

  json.NewEncoder(os.Stdout).Encode(spec)
}
