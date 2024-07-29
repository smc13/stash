package drivers

import (
	"context"
)

type Driver interface {
	// Initialize the cache driver if necessary, will be called by the Stash
	Init() error

	// Retrieve a value from the cache
	Get(ctx context.Context, key string) (*RawValue, error)
	// Store a value in the cache
	Put(ctx context.Context, raw RawValue) error
	// Store a value in the cache if the key does not exist / is expired
	Add(ctx context.Context, raw RawValue) error
	// Store a value in the cache forever
	Forever(ctx context.Context, raw RawValue) error
	// Remove a value from the cache
	Forget(ctx context.Context, key string) (bool, error)
	// Remove all values from the cache
	Flush(ctx context.Context) error
}
