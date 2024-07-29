package stash

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCacheMiss(t *testing.T) {
	var err error
	err = &CacheMissError{}
	assert.True(t, IsCacheMiss(err))

	err = errors.New("test")
	assert.False(t, IsCacheMiss(err))
}
