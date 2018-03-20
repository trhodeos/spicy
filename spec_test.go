package spicy

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParsingSimpleSpec(t *testing.T) {
	assert := assert.New(t)
	specStr := `
beginseg
  name "some_segment"
endseg
beginwave
  name "wave"
  include "some_segment"
endwave
`
	spec, err := ParseSpec(strings.NewReader(specStr))
	assert.Nil(err)
	assert.Equal(1, len(spec.Waves))

}
