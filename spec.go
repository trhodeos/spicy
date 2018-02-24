package spicy

import (
	"github.com/alecthomas/participle"
	"io"
)

type Constant struct {
	Symbol string `  @Ident`
	Int    int    `| @Int`
}

type Flag struct {
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
	Flags         []*Flag     `| @@ { @@ }`
	ConstantValue *Summand    `| @@`
	MaxSegment    *MaxSegment `| @@`
	MinSegment    *MinSegment `| @@`
}

type Statement struct {
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

type Statements struct {
	Statements []*Statement `{ @@ }`
}

type Segment struct {
	StatementList Statements `"beginseg" @@ "endseg"`
}

type Wave struct {
	StatementList Statements `"beginwave" @@ "endwave"`
}

type Spec struct {
	SegmentList []*Segment `{ @@ }`
	WaveList    []*Wave    `{ @@ }`
}

func ParseSpec(r io.Reader) (*Spec, error) {
	parser, err := participle.Build(&Spec{}, nil)
	if err != nil {
		return nil, err
	}

	spec := &Spec{}
	err = parser.Parse(r, spec)
	return spec, err
}
