package stash

import (
	"bytes"
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
func (s *Stash) Get(key string, value any) error {
	b, err := s.GetBytes(key)
	if err != nil {
		return err
	}

	return s.bytesToValue(b, value)
}

// Retrieve a value from the cache
func (s *Stash) GetBytes(key string) ([]byte, error) {
	raw, err := s.driver.Get(key)
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
func (s *Stash) Put(key string, value any, duration time.Duration) error {
	b, err := s.valueToBytes(value)
	if err != nil {
		return err
	}

	return s.PutBytes(key, b, duration)
}

// Store a pre-serialized value in the cache
func (s *Stash) PutBytes(key string, value []byte, duration time.Duration) error {
	return s.driver.Put(key, value, duration)
}

// Serialize a value and store it in the cache if the key does not exist / is expired
func (s *Stash) Add(key string, value any, duration time.Duration) error {
	b, err := s.valueToBytes(value)
	if err != nil {
		return err
	}

	return s.AddBytes(key, b, duration)
}

// Store a pre-serialized value in the cache if the key does not exist / is expired
func (s *Stash) AddBytes(key string, value []byte, duration time.Duration) error {
	return s.driver.Add(key, value, duration)
}

// Serialize a value and store it in the cache forever
func (s *Stash) Forever(key string, value any) error {
	b, err := s.valueToBytes(value)
	if err != nil {
		return err
	}

	return s.ForeverBytes(key, b)
}

// Store a pre-serialized value in the cache forever
func (s *Stash) ForeverBytes(key string, value []byte) error {
	return s.driver.Forever(key, value)
}

// Remove a value from the cache
func (s *Stash) Forget(key string) error {
	return s.driver.Forget(key)
}

// Remove all values from the cache, or only expired values
func (s *Stash) Flush() error {
	return s.driver.Flush()
}

// Check if a value exists in the cache
func (s *Stash) Has(key string) bool {
	raw, _ := s.driver.Get(key)
	return raw != nil
}

// Check if a value is missing from the cache
func (s *Stash) Missing(key string) bool {
	return !s.Has(key)
}

// Retrieve a value from the cache and then remove it
func (s *Stash) Pull(key string, value any) error {
	if err := s.Get(key, value); err != nil {
		return err
	}

	return s.Forget(key)
}

// Retrieve a value from the cache and then remove it
func (s *Stash) PullBytes(key string) ([]byte, error) {
	b, err := s.GetBytes(key)
	if err != nil {
		return nil, err
	}

	if err := s.Forget(key); err != nil {
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
