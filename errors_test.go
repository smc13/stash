package stash

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCacheMiss(t *testing.T) {
	assert.True(t, IsCacheMiss(cacheMissErr))
	assert.False(t, IsCacheMiss(errors.New("test")))
}
