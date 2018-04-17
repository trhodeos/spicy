package spicy

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestParsingSimpleSpec(t *testing.T) {
	assert := assert.New(t)
	specStr := `
beginseg
  name "obj"
  flags OBJECT
  include "some/file"
  address 0x12
endseg
beginseg
  name "raw"
  flags RAW
  include "some/file"
endseg
beginwave
  name "wave"
  include "obj"
  include "raw"
endwave
`
	spec, err := ParseSpec(strings.NewReader(specStr))
	assert.Nil(err)
	assert.Equal(1, len(spec.Waves))
	assert.Equal("wave", spec.Waves[0].Name)
	assert.Equal(1, len(spec.Waves[0].ObjectSegments))
	obj := spec.Waves[0].ObjectSegments[0]
	assert.Equal("obj", obj.Name)
	assert.Equal(uint64(0x12), obj.Positioning.Address)
	assert.Equal(1, len(spec.Waves[0].RawSegments))
	raw := spec.Waves[0].RawSegments[0]
	assert.Equal("raw", raw.Name)
}

func TestParsingIncludesWithRoot(t *testing.T) {
	assert := assert.New(t)
	specStr := `
beginseg
  name "some_segment"
  flags OBJECT
  include "some/file"
  include "$(ROOT)/some/file"
endseg
beginwave
  name "wave"
  include "some_segment"
endwave
`
	os.Setenv("ROOT", "parent")
	spec, err := ParseSpec(strings.NewReader(specStr))
	assert.Nil(err)
	assert.Equal(1, len(spec.Waves))
	assert.Equal(1, len(spec.Waves[0].ObjectSegments))
	assert.Equal(2, len(spec.Waves[0].ObjectSegments[0].Includes))
	assert.Equal("some/file", spec.Waves[0].ObjectSegments[0].Includes[0])
	assert.Equal("parent/some/file", spec.Waves[0].ObjectSegments[0].Includes[1])
}
