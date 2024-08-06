package stash

import (
	"encoding/json"
	"strconv"
	"strings"
	"unsafe"
)

type CacheResult struct {
	key string
	val string
	err error

	i *CacheItem
}

// Key returns the key of the cache result
func (c *CacheResult) Key() string {
	return c.key
}

// IsError returns true if the result is an error (including cache misses)
func (c *CacheResult) IsError() bool {
	return c.err != nil
}

// Error implements the error interface
func (c *CacheResult) Error() error {
	return c.err
}

func (c *CacheResult) AsString() (string, error) {
	return c.val, c.Error()
}

func (c *CacheResult) AsBytes() ([]byte, error) {
	str, err := c.AsString()
	if err != nil {
		return nil, err
	}

	return unsafe.Slice(unsafe.StringData(str), len(str)), nil
}

func (c *CacheResult) AsInt64() (int64, error) {
	v, err := c.AsString()
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(v, 10, 64)
}

func (c *CacheResult) AsJSON(val any) error {
	str, err := c.AsString()
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(strings.NewReader(str))
	return decoder.Decode(val)
}
