package stash

import (
	"errors"
	"testing"

	"github.com/smc13/stash/drivers"
	"github.com/stretchr/testify/assert"
)

func TestIsStashError(t *testing.T) {
	var err error
	err = &drivers.CacheMissError{}
	assert.True(t, IsStashError(err))

	err = errors.New("test")
	assert.False(t, IsStashError(err))
}

func TestIsCacheError(t *testing.T) {
	var err error
	err = &drivers.CacheMissError{}
	assert.True(t, IsCacheError(err))

	err = errors.New("test")
	assert.False(t, IsCacheError(err))
}
