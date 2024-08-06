package stash

import (
	"time"
)

type CacheItem struct {
	// The unique key for the cache item
	Key string
	// Value of the cache item as a string
	Value string
	// The time after which the cache item will expire
	Expires time.Time
}

// Returns true if the cache item has expired
func (ci *CacheItem) IsExpired() bool {
	return ci.Expires.Before(time.Now())
}
