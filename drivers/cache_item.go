package drivers

import (
	"time"
)

type CacheItem interface {
	Key() string
	Value() string
	Expires() time.Time
}
