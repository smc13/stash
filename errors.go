package stash

import "errors"

var cacheMissErr = errors.New("cache miss")

// Returns true if the error is a CacheMissError
func IsCacheMiss(err error) bool {
	return errors.Is(err, cacheMissErr)
}
