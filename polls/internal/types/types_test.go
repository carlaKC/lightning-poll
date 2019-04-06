package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExtreme(t *testing.T) {
	greaterThan := func(challenger, existing int64) bool {
		return challenger > existing
	}
	lessThan := func(challenger, existing int64) bool {
		return challenger < existing
	}

	// poll outputs tie
	m := map[int64]int64{
		1: 10,
		2: 10,
	}

	ok := getExtreme(m, 1, greaterThan)
	assert.True(t, ok)
	ok = getExtreme(m, 2, greaterThan)
	assert.True(t, ok)

	ok = getExtreme(m, 1, lessThan)
	assert.True(t, ok)
	ok = getExtreme(m, 2, lessThan)
	assert.True(t, ok)

	// poll outputs differ
	m = map[int64]int64{
		1: 0,
		2: 10,
	}

	ok = getExtreme(m, 1, greaterThan)
	assert.False(t, ok)
	ok = getExtreme(m, 2, greaterThan)
	assert.True(t, ok)

	ok = getExtreme(m, 1, lessThan)
	assert.True(t, ok)
	ok = getExtreme(m, 2, lessThan)
	assert.False(t, ok)
}
