package spicy

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/alecthomas/participle"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Constant struct {
	Symbol string `  @Ident`
	Int    uint64 `| @Int`
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
	Int           uint64      `| @Int`
	Flags         []*FlagAst  `| @@ { @@ }`
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
	Waves    []*WaveAst    `{ @@ }`
}

type Flags struct {
	Object bool
	Boot   bool
	Raw    bool
}

type Positioning struct {
	AfterSegment    string
	AfterMinSegment [2]string
	AfterMaxSegment [2]string
	Address         uint64
}

type StackInfo struct {
	Start  string
	Offset uint64
}

type Segment struct {
	Name        string
	Includes    []string
	StackInfo   *StackInfo
	Positioning Positioning
	Entry       *string
	MaxSize     uint64
	Align       uint64
	Flags       Flags
}

type Wave struct {
	Name           string
	ObjectSegments []*Segment
	RawSegments    []*Segment
}

type Spec struct {
	Waves []*Wave
}

func convertSegmentAst(s *SegmentAst) (*Segment, error) {
	seg := &Segment{}
	for _, statement := range s.Statements {
		switch statement.Name {
		case "name":
			seg.Name = statement.Value.String
			break
		case "address":
			seg.Positioning.Address = SignExtend(statement.Value.Int)
			break
		case "after":
			if statement.Value.String != "" {
				seg.Positioning.AfterSegment = statement.Value.String
			} else if statement.Value.MinSegment != nil {
				seg.Positioning.AfterMinSegment = [2]string{statement.Value.MinSegment.First, statement.Value.MinSegment.Second}
			} else if statement.Value.MaxSegment != nil {
				seg.Positioning.AfterMaxSegment = [2]string{statement.Value.MaxSegment.First, statement.Value.MaxSegment.Second}
			} else {
				return nil, errors.New("No value found in 'after' statement")
			}
			break
		case "include":
			// Hacky way of moving $(var) -> $var
			replaced := strings.Replace(statement.Value.String, "$(", "$", -1)
			replaced = strings.Replace(replaced, ")", "", -1)
			replaced = filepath.Clean(os.ExpandEnv(replaced))
			seg.Includes = append(seg.Includes, replaced)
			break
		case "maxsize":
			seg.MaxSize = statement.Value.Int
			break
		case "align":
			seg.Align = statement.Value.Int
			break
		case "flags":
			for _, f := range statement.Value.Flags {
				if f.Boot {
					seg.Flags.Boot = true
				} else if f.Object {
					seg.Flags.Object = true
				} else if f.Raw {
					seg.Flags.Raw = true
				}
			}
			break
		case "number":
			// Don't do anything, as we don't really care here.
			// All that matters for code is the rom address.
			break
		case "entry":
			seg.Entry = &statement.Value.ConstantValue.Lhs.Symbol
			break
		case "stack":
			seg.StackInfo = &StackInfo{}
			if statement.Value.ConstantValue.Lhs.Symbol != "" {
				seg.StackInfo.Start = statement.Value.ConstantValue.Lhs.Symbol
			} else {
				seg.StackInfo.Start = string(statement.Value.ConstantValue.Lhs.Int)
			}
			if statement.Value.ConstantValue.Rhs.Int != 0 {
				seg.StackInfo.Offset = statement.Value.ConstantValue.Rhs.Int
			}
			break
		default:
			return nil, errors.New(fmt.Sprintf("Unknown name %s", statement.Name))
		}
	}
	return seg, nil
}

func convertWaveAst(s *WaveAst, segments map[string]*Segment) (*Wave, error) {
	out := &Wave{}
	for _, statement := range s.Statements {
		switch statement.Name {
		case "name":
			out.Name = statement.Value.String
			break
		case "include":
			seg := segments[statement.Value.String]
			if seg.Flags.Object {
				out.ObjectSegments = append(out.ObjectSegments, seg)
			} else if seg.Flags.Raw {
				out.RawSegments = append(out.RawSegments, seg)
			}
			break
		default:
			return nil, errors.New(fmt.Sprintf("Unknown name %s", statement.Name))
		}
	}
	return out, nil
}

func (w *Wave) updateWithConstants() {
	for _, seg := range w.ObjectSegments {
		if seg.Flags.Boot && seg.Positioning.Address == 0 {
			seg.Positioning.Address = SignExtend(0x80000450)
		}
	}
}

func convertAstToSpec(s SpecAst) (*Spec, error) {
	out := &Spec{}
	segments := map[string]*Segment{}
	for _, segAst := range s.Segments {
		seg, err := convertSegmentAst(segAst)
		if err != nil {
			return nil, err
		}
		segments[seg.Name] = seg
	}
	for _, waveAst := range s.Waves {
		wave, err := convertWaveAst(waveAst, segments)
		if err != nil {
			return nil, err
		}
		wave.updateWithConstants()
		err = wave.checkValidity()
		if err != nil {
			return nil, err
		}
		out.Waves = append(out.Waves, wave)
	}

	return out, nil
}

func PreprocessSpec(file io.Reader, gcc Runner, includeFlags []string, defineFlags []string, undefineFlags []string) (io.Reader, error) {
	args := []string{"-P", "-E", "-U_LANGUAGE_C", "-D_LANGUAGE_MAKEROM", "-"}
	for _, include := range includeFlags {
		args = append(args, fmt.Sprintf("-I%s", include))
	}
	for _, define := range defineFlags {
		args = append(args, fmt.Sprintf("-D%s", define))
	}
	for _, undefine := range undefineFlags {
		args = append(args, fmt.Sprintf("-U%s", undefine))
	}

	return gcc.Run(file, args)
}

func ParseSpec(r io.Reader) (*Spec, error) {
	log.Infof("Parsing spec")
	parser, err := participle.Build(&SpecAst{}, nil)
	if err != nil {
		return nil, err
	}

	specAst := &SpecAst{}
	err = parser.Parse(r, specAst)
	if err != nil {
		return nil, err
	}
	out, err := convertAstToSpec(*specAst)
	if err == nil {
		log.Debugf("Parsed: %v", out)
	}
	for _, w := range out.Waves {
		w.correctOrdering()
	}
	return out, err
}

func (w *Wave) checkValidity() error {
	for _, seg := range w.ObjectSegments {
		numSet := 0
		if seg.Name == "" {
			return errors.New("Name must be non-empty.")
		}
		if seg.Flags.Boot && seg.StackInfo == nil {
			return errors.New("Boot segments must have stack info specified.")
		}
		if seg.Flags.Boot && seg.Entry == nil {
			return errors.New("Boot segments must have entry point specified.")
		}
		if seg.Positioning.Address > 0 {
			numSet++
		}
		if seg.Positioning.AfterSegment != "" {
			numSet++
		}
		if seg.Positioning.AfterMinSegment[0] != "" {
			numSet++
		}
		if seg.Positioning.AfterMaxSegment[0] != "" {
			numSet++
		}
		if numSet > 1 {
			return errors.New(fmt.Sprintf("Too many addressing sections specified in segment %s.", seg.Name))
		}
	}
	// Per-spec checks
	// Wave checks
	return nil
}

func findElement(l *list.List, name string) *list.Element {
	for e := l.Front(); e != nil; e = e.Next() {
		original, _ := e.Value.(*Segment)
		if original.Name == name {
			return e
		}
	}
	return nil
}

func (w *Wave) correctOrdering() {
	// TODO(trhodeos): Make this actually correct. This just places any
	// explicitly addressed segments before any 'after.*' segments. If
	// 'after' segments depend on other 'after' segments, this'll break.
	l := list.New()
	for _, seg := range w.ObjectSegments {
		if seg.Positioning.Address != 0 {
			l.PushBack(seg)
		}
	}
	for _, seg := range w.ObjectSegments {
		if seg.Positioning.Address == 0 {
			e := findElement(l, seg.Positioning.AfterSegment)
			l.InsertAfter(seg, e)
		}
	}

	newSegments := make([]*Segment, 0)
	for e := l.Front(); e != nil; e = e.Next() {
		original, _ := e.Value.(*Segment)
		newSegments = append(newSegments, original)
	}
	w.ObjectSegments = newSegments
}

func (w *Wave) GetBootSegment() *Segment {
	for _, seg := range w.ObjectSegments {
		if seg.Flags.Boot {
			return seg
		}
	}
	return nil
}
