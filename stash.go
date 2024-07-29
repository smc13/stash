package stash

import (
	"bytes"
	"context"
	"encoding/gob"
	"log/slog"
	"time"

	"github.com/smc13/stash/drivers"
)

type Stash struct {
	driver drivers.Driver
	logger slog.Logger
}

// Creates a new Stash instance with the given driver
func New(driver drivers.Driver) (*Stash, error) {
	if err := driver.Init(); err != nil {
		return nil, err
	}

	return &Stash{driver: driver}, nil
}

func (s *Stash) WithLogger(logger slog.Logger) *Stash {
	s.logger = logger
	return s
}

// Retrieve a value from the cache and deserialize it
func (s *Stash) Get(ctx context.Context, key string, value any) error {
	b, err := s.GetBytes(ctx, key)
	if err != nil {
		return err
	}

	return s.bytesToValue(b, value)
}

// Retrieve a value from the cache
func (s *Stash) GetBytes(ctx context.Context, key string) ([]byte, error) {
	raw, err := s.driver.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if raw == nil || raw.IsExpired() {
		var at *time.Time
		if raw != nil {
			at = &raw.Expires
		}

		return nil, &CacheMissError{key, at}
	}

	return raw.Value, nil
}

// Serialize a value and store it in the cache
func (s *Stash) Put(ctx context.Context, key string, value any, duration time.Duration) error {
	b, err := s.valueToBytes(value)
	if err != nil {
		return err
	}

	return s.PutBytes(ctx, key, b, duration)
}

// Store a pre-serialized value in the cache
func (s *Stash) PutBytes(ctx context.Context, key string, value []byte, duration time.Duration) error {
	return s.driver.Put(ctx, drivers.RawValueFromBytes(key, value, s.durationToTime(duration)))
}

// Serialize a value and store it in the cache if the key does not exist / is expired
func (s *Stash) Add(ctx context.Context, key string, value any, duration time.Duration) error {
	b, err := s.valueToBytes(value)
	if err != nil {
		return err
	}

	return s.AddBytes(ctx, key, b, duration)
}

// Store a pre-serialized value in the cache if the key does not exist / is expired
func (s *Stash) AddBytes(ctx context.Context, key string, value []byte, duration time.Duration) error {
	return s.driver.Add(ctx, drivers.RawValueFromBytes(key, value, s.durationToTime(duration)))
}

// Serialize a value and store it in the cache forever
func (s *Stash) Forever(ctx context.Context, key string, value any) error {
	b, err := s.valueToBytes(value)
	if err != nil {
		return err
	}

	return s.ForeverBytes(ctx, key, b)
}

// Store a pre-serialized value in the cache forever
func (s *Stash) ForeverBytes(ctx context.Context, key string, value []byte) error {
	return s.driver.Forever(ctx, drivers.RawValueFromBytes(key, value, s.durationToTime(0)))
}

// Remove a value from the cache
func (s *Stash) Forget(ctx context.Context, key string) (bool, error) {
	return s.driver.Forget(ctx, key)
}

// Remove all values from the cache, or only expired values
func (s *Stash) Flush(ctx context.Context) error {
	return s.driver.Flush(ctx)
}

// Check if a value exists in the cache
func (s *Stash) Has(ctx context.Context, key string) bool {
	raw, _ := s.driver.Get(ctx, key)
	return raw != nil
}

// Check if a value is missing from the cache
func (s *Stash) Missing(ctx context.Context, key string) bool {
	return !s.Has(ctx, key)
}

// Retrieve a value from the cache and then remove it
func (s *Stash) Pull(ctx context.Context, key string, value any) (bool, error) {
	if err := s.Get(ctx, key, value); err != nil {
		return false, err
	}

	return s.Forget(ctx, key)
}

// Retrieve a value from the cache and then remove it
func (s *Stash) PullBytes(ctx context.Context, key string) ([]byte, error) {
	b, err := s.GetBytes(ctx, key)
	if err != nil {
		return nil, err
	}

	if _, err := s.Forget(ctx, key); err != nil {
		return nil, err
	}

	return b, nil
}

// Convert an unknown value to a byte slice
func (s *Stash) valueToBytes(value any) ([]byte, error) {
	if stashable, ok := value.(Stashable); ok {
		return stashable.ToStash()
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Convert a byte slice back to an unknown value
func (s *Stash) bytesToValue(b []byte, value any) error {
	if stashable, ok := value.(Stashable); ok {
		return stashable.FromStash(b)
	}

	dec := gob.NewDecoder(bytes.NewReader(b))
	return dec.Decode(value)
}

func (s *Stash) durationToTime(d time.Duration) time.Time {
	if d == 0 {
		return time.Time{}
	}

	return time.Now().Add(d)
}
