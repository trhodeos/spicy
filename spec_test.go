package spicy

import (
	"strings"
	"testing"
)

func TestParsingSimpleSpec(t *testing.T) {
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
	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}
	if len(spec.Waves) != 1 {
		t.Errorf("Expected 1 wave, found %d", len(spec.Waves))
	}

}
