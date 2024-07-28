package stash

import (
	"fmt"
	"time"
)

type CacheMissError struct {
	// The key that was attempted to be retrieved
	Key    string
	// An expiry time if the key was found but expired
	Expiry *time.Time
}

func (e *CacheMissError) Error() string {
	return fmt.Sprintf("cache miss for key %s", e.Key)
}

// Returns true if the error is a CacheMissError
func IsCacheMiss(err error) bool {
	_, ok := err.(*CacheMissError)
	return ok
}
