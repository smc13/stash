package stash

import (
	"fmt"
	"time"
)

type StashItem struct {
	// The unique key for the cache item
	key string
	// Value of the cache item, encodable to a string
	value any
	// The time after which the cache item will expire
	expires time.Time
}

func NewStashItem(key string, value any, expires time.Time) (*StashItem, error) {
	return &StashItem{key: key, value: value, expires: expires}, nil
}

func (si *StashItem) Key() string        { return si.key }
func (si *StashItem) Expires() time.Time { return si.expires }
func (si *StashItem) RawValue() any      { return si.value }
func (si *StashItem) Value() string {
	var valString string
	switch si.value.(type) {
	case string:
		valString = si.value.(string)
	case fmt.Stringer:
		valString = si.value.(fmt.Stringer).String()
	default:
		valString = JSON(si.value)
	}

	return valString
}
