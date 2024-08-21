package stash

import (
	"fmt"
	"time"

	"github.com/smc13/stash/drivers"
)

type Stashable interface {
	ToStash() drivers.CacheItem
}

type StashItem struct {
	// The unique key for the cache item
	Key string
	// Value of the cache item, encodable to a string
	Value any
	// The time after which the cache item will expire
	Expires time.Time
}

type stashCacheItem struct {
	key     string
	value   string
	expires time.Time
}

func (sci *stashCacheItem) Key() string        { return sci.key }
func (sci *stashCacheItem) Value() string      { return sci.value }
func (sci *stashCacheItem) Expires() time.Time { return sci.expires }

func (si *StashItem) ToStash() drivers.CacheItem {
	var valString string
	switch si.Value.(type) {
	case string:
		valString = si.Value.(string)
	case fmt.Stringer:
		valString = si.Value.(fmt.Stringer).String()
	default:
		valString = JSON(si.Value)
	}

	return &stashCacheItem{
		key:     si.Key,
		value:   valString,
		expires: si.Expires,
	}
}
