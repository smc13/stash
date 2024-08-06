package stash

import (
	"fmt"
	"time"
)

type Stashable interface {
	ToStash() *CacheItem
}

type StashItem struct {
	// The unique key for the cache item
	Key string
	// Value of the cache item, encodable to a string
	Value any
	// The time after which the cache item will expire
	Expires time.Time
}

func (si *StashItem) ToStash() *CacheItem {
	var valString string
	switch si.Value.(type) {
	case string:
		valString = si.Value.(string)
	case fmt.Stringer:
		valString = si.Value.(fmt.Stringer).String()
	default:
		valString = JSON(si.Value)
	}

	return &CacheItem{
		Key:     si.Key,
		Value:   valString,
		Expires: si.Expires,
	}
}
