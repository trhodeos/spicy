package spicy

import (
	"github.com/alecthomas/participle"
	"io"
        "errors"
        "fmt"
        "text/template"
        "os"
)

type Constant struct {
	Symbol string `  @Ident`
	Int    int    `| @Int`
}

type FlagAst struct {
	Boot   bool `  @"BOOT"`
	Object bool `| @"OBJECT"`
	Raw    bool `| @"RAW"`
}

type Summand struct {
	Lhs *Constant ` @@`
	Op  string    `[ @("+" | "-")`
	Rhs *Constant ` @@ ]`
}

type MaxSegment struct {
	First  string `"max[" @String ","`
	Second string `    @String "]"`
}

type MinSegment struct {
	First  string `"min[" @String ","`
	Second string `       @String "]"`
}

// Only one of these values will be set.
type Value struct {
	String        string      `  @String`
	Int           int    `| @Int`
	Flags         []*FlagAst     `| @@ { @@ }`
	ConstantValue *Summand    `| @@`
	MaxSegment    *MaxSegment `| @@`
	MinSegment    *MinSegment `| @@`
}

type StatementAst struct {
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
	// I tried using @Ident here, but the parser was greedily taking 'endseg' as name.
	// By explicitly listing all known names here, we limit the search space.
	Name  string `@("name" | "address" | "after" | "include" | "maxsize" | "align" | "flags" | "number" | "entry" | "stack")`
	Value Value  `@@`
}

type SegmentAst struct {
	Statements []*StatementAst `"beginseg" { @@ } "endseg"`
}

type WaveAst struct {
	Statements []*StatementAst `"beginwave" { @@ } "endwave"`
}

type SpecAst struct {
	Segments []*SegmentAst `{ @@ }`
	Waves []*WaveAst    `{ @@ }`
}

type Flags struct {
  Object bool
  Boot bool
  Raw bool
}

type Positioning struct {
  AfterSegment string
  AfterMinSegment [2]string
  AfterMaxSegment [2]string
  Address int
  Number int
}

type StackInfo struct {
  StartSymbol string
  StartAddress int
  Offset int
}

type Segment struct {
  Name string
  Includes []string
  StackInfo StackInfo
  Positioning Positioning
  Entry string
  MaxSize int
  Align int
  Flags Flags
}

type Wave struct {
  Name string
  Includes []string
}

type Spec struct {
  Segments []*Segment
  Waves []*Wave
}

func convertSegmentAst(s *SegmentAst) (*Segment, error) {
  seg := &Segment{}
  for _, statement := range(s.Statements) {
    switch (statement.Name) {
      case "name":
        seg.Name = statement.Value.String
        break
      case "address":
        seg.Positioning.Address = statement.Value.Int
        break
      case "after":
        if statement.Value.String != "" {
          seg.Positioning.AfterSegment = statement.Value.String
        } else if statement.Value.MinSegment.First != "" {
          seg.Positioning.AfterMinSegment = [2]string{statement.Value.MinSegment.First, statement.Value.MinSegment.Second}
        } else if statement.Value.MaxSegment.First != "" {
          seg.Positioning.AfterMaxSegment = [2]string{statement.Value.MaxSegment.First, statement.Value.MaxSegment.Second}
        } else {
          return nil, errors.New("some error")
        }
        break
      case "include":
        seg.Includes = append(seg.Includes, statement.Value.String)
        break
      case "maxsize":
        seg.MaxSize = statement.Value.Int
        break
      case "align":
        seg.Align = statement.Value.Int
        break
      case "flags":
        for _, f := range(statement.Value.Flags) {
          if (f.Boot) {
            seg.Flags.Boot = true
          } else if (f.Object) {
            seg.Flags.Object = true
          } else if (f.Raw) {
            seg.Flags.Raw = true
          }
        }
        break
      case "number":
        seg.Positioning.Number = statement.Value.Int
        break
      case "entry":
        seg.Entry = statement.Value.ConstantValue.Lhs.Symbol
        break
      case "stack":
        if (statement.Value.ConstantValue.Lhs.Symbol != "") {
          seg.StackInfo.StartSymbol = statement.Value.ConstantValue.Lhs.Symbol
        } else {
          seg.StackInfo.StartAddress = statement.Value.ConstantValue.Lhs.Int
        }
        if (statement.Value.ConstantValue.Rhs.Int != 0) {
          seg.StackInfo.Offset = statement.Value.ConstantValue.Rhs.Int
        }
        break
      default:
        return nil, errors.New(fmt.Sprintf("Unknown name %s", statement.Name))
    }
  }
  return seg, nil
}

func convertWaveAst(s *WaveAst) (*Wave, error) {
  out := &Wave{}
  for _, statement := range(s.Statements) {
    switch (statement.Name) {
      case "name":
        out.Name = statement.Value.String
        break
      case "include":
        out.Includes = append(out.Includes, statement.Value.String)
        break
      default:
        return nil, errors.New(fmt.Sprintf("Unknown name %s", statement.Name))
    }
  }
  return out, nil
}

func convertAstToSpec(s SpecAst) (*Spec, error) {
  out := &Spec{}
  for _, segAst := range(s.Segments) {
    seg, err := convertSegmentAst(segAst)
    if err != nil {
      return nil, err
    }
    out.Segments = append(out.Segments, seg)
  }
  for _, waveAst := range(s.Waves) {
    wave, err := convertWaveAst(waveAst)
    if err != nil {
      return nil, err
    }
    out.Waves = append(out.Waves, wave)
  }

  return out, nil
}

func ParseSpec(r io.Reader) (*Spec, error) {
	parser, err := participle.Build(&SpecAst{}, nil)
	if err != nil {
		return nil, err
	}

	specAst := &SpecAst{}
	err = parser.Parse(r, specAst)
        if err != nil {
          return nil, err
        }
        return convertAstToSpec(*specAst)
}

func (s *Spec) GenerateLdScript() (string, error) {
  t := `
MEMORY {
  {{range .Segments}}
  {{.Name}}.RAM (RX) : ORIGIN = 0x80000450, LENGTH = 0x400000
  {{.Name}}.bss.RAM (RW) : ORIGIN = 0x80000450, LENGTH = 0x400000
  {{end}}
}
`
  tmpl, err := template.New("test").Parse(t)
  if err != nil {
    return "", err
  }
  err = tmpl.Execute(os.Stdout, s)
  return "", err
}
