package drivers

import (
	"time"
)

type Driver interface {
	// Initialize the cache driver if necessary, will be called by the Stash
	Init() error

	// Retrieve a value from the cache
	Get(key string) (*RawValue, error)
	// Store a value in the cache
	Put(key string, value []byte, expires time.Duration) error
	// Store a value in the cache if the key does not exist / is expired
	Add(key string, value []byte, expires time.Duration) error
	// Store a value in the cache forever
	Forever(key string, value []byte) error
	// Remove a value from the cache
	Forget(key string) error
	// Remove all values from the cache
	Flush() error
}

type RawValue struct {
	Key     string
	Value   []byte
	Expires time.Time
}

func (rv *RawValue) IsExpired() bool {
	return rv.Expires.Before(time.Now())
}
