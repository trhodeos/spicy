package main

import (
  "encoding/json"

  "os"

  "github.com/alecthomas/participle"
)


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
  SegmentName string      `"name" @String |`
  Address     string    `"address" @String |`
  After       string      `"after" @String |`
  Include     string     `"include" @String |`
  MaxSize     string     `"maxsize" @String |`
  Align       string     `"align" @String |`
  Flags       string     `"flags" @String |`
  Number      string     `"number" @String |`
  entry string     `"entry" @String |`
  stack string     `"stack" @String`
}

type Segment struct {
  SegmentStatementList []*SegmentStatement `"beginseg" { @@ } "endseg"`
}

type Wave struct {
  WaveStatementList []*string `"beginwave" { @String } "endwave"`
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
