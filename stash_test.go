package stash

import (
	"context"
	"testing"
	"time"

	"github.com/smc13/stash/drivers"
	"github.com/stretchr/testify/assert"
)

type user struct {
	ID   int
	Name string
}

func stashValues(t *testing.T, ctx context.Context, stash *Stash) {
	assert.NoError(t, stash.Put(ctx, "bytes", []byte("bytes"), 5*time.Minute))
	assert.NoError(t, stash.Put(ctx, "string", "string", 5*time.Minute))
	assert.NoError(t, stash.Put(ctx, "int", 1, 5*time.Minute))
	assert.NoError(t, stash.Put(ctx, "struct", user{1, "test"}, 5*time.Minute))
}

func TestPut(t *testing.T) {
	tests := []struct {
		name   string
		driver drivers.Driver
	}{
		{"file", drivers.NewFileDriver(t.TempDir())},
		{"memory", drivers.NewMemoryDriver()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			stash, _ := New(tt.driver)

			stashValues(t, ctx, stash)
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name   string
		driver drivers.Driver
	}{
		{"file", drivers.NewFileDriver(t.TempDir())},
		{"memory", drivers.NewMemoryDriver()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			stash, _ := New(tt.driver)

			stashValues(t, ctx, stash)

			var b []byte
			assert.NoError(t, stash.Get(ctx, "bytes", &b))
			assert.Equal(t, []byte("bytes"), b)

			var s string
			assert.NoError(t, stash.Get(ctx, "string", &s))
			assert.Equal(t, "string", s)

			var i int
			assert.NoError(t, stash.Get(ctx, "int", &i))
			assert.Equal(t, 1, i)

			var u user
			assert.NoError(t, stash.Get(ctx, "struct", &u))
			assert.Equal(t, user{1, "test"}, u)
		})
	}
}
