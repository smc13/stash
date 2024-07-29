package drivers

import (
	"encoding/hex"
	"time"
)

type RawValue struct {
	Key     string
	Value   []byte
	Expires time.Time
}

func (rv *RawValue) IsExpired() bool {
	return rv.Expires.Before(time.Now())
}

func (rv *RawValue) AsString() string {
	return hex.EncodeToString(rv.Value)
}

// create a raw value from a byte slice
// typically used when storing a value in a cache driver
func RawValueFromBytes(key string, value []byte, expires time.Time) RawValue {
	return RawValue{
		Key:     key,
		Value:   value,
		Expires: expires,
	}
}

// create a raw value from a hex encoded string
// typically used when returning from a cache driver
func RawValueFromString(key string, value string, expires time.Time) (*RawValue, error) {
	b, err := hex.DecodeString(value)
	if err != nil {
		return nil, err
	}

	return &RawValue{
		Key:     key,
		Value:   b,
		Expires: expires,
	}, nil
}
