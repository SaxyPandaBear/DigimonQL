package db

import (
	"testing"

	"github.com/saxypandabear/digimonql/graph/model"
	"github.com/stretchr/testify/assert"
)

// Verify that the struct can be coerced into the interface
func TestConfirmsLocalToInterface(t *testing.T) {
	var _ DigimonRepository = &LocalDigimonRepository{}
	t.SkipNow()
}

func TestLocalFilterMatching(t *testing.T) {
	val := "foo"
	f := model.Filter{
		Name: &val,
	}

	d := model.Digimon{
		Name: val,
	}

	assert.True(t, matchesFilter(&d, &f))

	d.Name = "something else"
	assert.False(t, matchesFilter(&d, &f))

	f.Moves = []string{"a", "b"}
	d.Moves = []string{"a", "b", "c"}
	d.Name = val // reset

	assert.True(t, matchesFilter(&d, &f))

	d.Moves = []string{"b"} // tests for && on the filter
	assert.False(t, matchesFilter(&d, &f))

	f.Moves = nil // reset
	isTrue := true
	f.IsXAntibody = &isTrue
	d.IsXAntibody = true

	assert.True(t, matchesFilter(&d, &f))
	d.IsXAntibody = false
	assert.False(t, matchesFilter(&d, &f))
}
