package stash

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
)

type CacheResult struct {
	*strings.Reader

	key string
	val string
}

func newCacheResult(key, val string) *CacheResult {
	log.Println(val)
	return &CacheResult{
		Reader: strings.NewReader(val),
		key:    key,
		val:    val,
	}
}

// Key returns the key of the cache result
func (c *CacheResult) Key() string {
	return c.key
}

func AsString(c *CacheResult) string {
	return c.val
}

func As[t any](c *CacheResult) (t, error) {
	var zero t
	var pzero any = &zero

	var a any
	var err error
	switch pzero.(type) {
	case string:
		a = AsString(c)
	case int:
		a, err = strconv.Atoi(c.val)
	case int64:
		a, err = strconv.ParseInt(c.val, 10, 64)
	case uint64:
		a, err = strconv.ParseUint(c.val, 10, 64)

	default:
		decoder := json.NewDecoder(c)
		if err = decoder.Decode(&zero); err != nil {
			return zero, err
		}

		a = zero
	}

	log.Println(a)
	return a.(t), nil
}

func AsDefault[t any](c *CacheResult, cb func() (t, error)) (t, error) {
	if c == nil {
		return cb()
	}

	return As[t](c)
}
