package stash

import (
	"context"
	"log/slog"

	"errors"

	"github.com/smc13/stash/drivers"
)

type Stash struct {
	driver drivers.Driver

	logger *slog.Logger
	errKey string
}

// Creates a new Stash instance with the given driver
func New(driver drivers.Driver, opts ...StashOption) (*Stash, error) {
	if err := driver.Init(); err != nil {
		return nil, err
	}

	stash := &Stash{driver: driver, logger: slog.Default(), errKey: "error"}
	for _, opt := range opts {
		if err := opt(stash); err != nil {
			return nil, err
		}
	}

	return stash, nil
}

// Get a value from the cache
func (s *Stash) Get(ctx context.Context, key string) *CacheResult {
	item, err := s.driver.Get(ctx, key)
	if err != nil {
		s.logger.ErrorContext(ctx, "error retriving key from cache", slog.String("key", key), slog.String(s.errKey, err.Error()))
		return &CacheResult{key, "", errors.Join(err, errors.New("cache get error"))}
	}

	if item == nil {
		slog.DebugContext(ctx, "cache miss", slog.String("key", key))
		return &CacheResult{key, "", cacheMissErr}
	}

	slog.DebugContext(ctx, "cache hit", slog.String("key", key))
	return &CacheResult{key, *item, nil}
}

// Put a value in the cache
func (s *Stash) Put(ctx context.Context, item drivers.CacheItem) error {
	if item == nil {
		return errors.New("no cache item to store")
	}

	var err error
	if item.Expires().IsZero() {
		err = s.driver.Forever(ctx, item)
	} else {
		err = s.driver.Put(ctx, item)
	}

	if err != nil {
		s.logger.ErrorContext(ctx, "error while storing key in cache", slog.String("key", item.Key()), slog.String(s.errKey, err.Error()))
	}
	return err
}

// Add a value to the cache if it does not exist
func (s *Stash) Add(ctx context.Context, item drivers.CacheItem) error {
	err := s.driver.Add(ctx, item)
	if err != nil {
		s.logger.ErrorContext(ctx, "error while adding key to cache", slog.String("key", item.Key()), slog.String(s.errKey, err.Error()))
	}

	return err
}

// Forget removes a value from the cache
func (s *Stash) Forget(ctx context.Context, key string) (bool, error) {
	found, err := s.driver.Forget(ctx, key)
	if err != nil {
		s.logger.ErrorContext(ctx, "error while removing key from cache", slog.String("key", key), slog.String(s.errKey, err.Error()))
	}

	return found, err
}

// Flush removes all values from the cache
// warning: no gurantee that prefix will be respected
func (s *Stash) Flush(ctx context.Context) error {
	err := s.driver.Flush(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "error while flushing cache", slog.String(s.errKey, err.Error()))
	}

	return err
}

// Has checks if a value exists in the cache
func (s *Stash) Has(ctx context.Context, key string) bool {
	raw, _ := s.driver.Get(ctx, key)
	return raw != nil
}

// Missing checks if a value does not exist in the cache
func (s *Stash) Missing(ctx context.Context, key string) bool {
	return !s.Has(ctx, key)
}

// Pull gets an item from the cache, then removes it
func (s *Stash) Pull(ctx context.Context, key string) *CacheResult {
	result := s.Get(ctx, key)
	if result.IsError() {
		return result
	}

	_, err := s.Forget(ctx, key)
	if err != nil {
		result.err = errors.Join(err, errors.New("cache pull error"))
	}

	return result
}
