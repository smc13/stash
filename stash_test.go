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
	assert.NoError(t, stash.Put(ctx, &StashItem{"bytes", BinaryString([]byte("bytes")), time.Now().Add(5 * time.Minute)}))
	assert.NoError(t, stash.Put(ctx, &StashItem{"string", "string", time.Now().Add(5 * time.Minute)}))
	assert.NoError(t, stash.Put(ctx, &StashItem{"int", 1, time.Now().Add(5 * time.Minute)}))
	assert.NoError(t, stash.Put(ctx, &StashItem{"struct", &user{1, "Hello world"}, time.Now().Add(5 * time.Minute)}))
	assert.NoError(t, stash.Put(ctx, &StashItem{"forever", "forever", time.Time{}}))
}

func TestPut(t *testing.T) {
	tests := []struct {
		name   string
		driver drivers.Driver
	}{
		{"file", drivers.NewFileDriver(t.TempDir()).Prefix("test")},
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
		{"file", drivers.NewFileDriver("./test")},
		{"memory", drivers.NewMemoryDriver()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			stash, _ := New(tt.driver)

			stashValues(t, ctx, stash)

			res, err := stash.Get(ctx, "bytes")
			assert.NoError(t, err)

			res, err = stash.Get(ctx, "string")
			assert.NoError(t, err)
			s := AsString(res)
			assert.Equal(t, "string", s)

			res, err = stash.Get(ctx, "int")
			assert.NoError(t, err)
			i, err := As[int64](res)
			assert.NoError(t, err)
			assert.Equal(t, int64(1), i)

			res, err = stash.Get(ctx, "struct")
			assert.NoError(t, err)
			u, err := As[user](res)
			assert.NoError(t, err)
			assert.Equal(t, user{1, "Hello world"}, u)
		})
	}
}
