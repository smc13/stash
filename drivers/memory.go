package drivers

import (
	"sync"
	"time"
)

type memoryDriver struct {
	values map[string]*RawValue
	mutx   sync.Mutex
}

func NewMemoryDriver() Driver {
	return &memoryDriver{
		values: make(map[string]*RawValue),
		mutx:   sync.Mutex{},
	}
}

func (d *memoryDriver) Init() error { return nil }

func (d *memoryDriver) Add(key string, value []byte, expires time.Duration) error {
	d.mutx.Lock()
	defer d.mutx.Unlock()

	raw, found := d.values[key]
	if found && !raw.IsExpired() {
		return nil
	}

	d.values[key] = &RawValue{
		Key:     key,
		Value:   value,
		Expires: time.Now().Add(expires),
	}

	return nil
}

func (d *memoryDriver) Flush() error {
	d.mutx.Lock()
	defer d.mutx.Unlock()

	d.values = make(map[string]*RawValue)
	return nil
}

func (d *memoryDriver) Forever(key string, value []byte) error {
	return d.Put(key, value, 999999999*time.Second)
}

func (d *memoryDriver) Forget(key string) error {
	d.mutx.Lock()
	defer d.mutx.Unlock()

	delete(d.values, key)
	return nil
}

func (d *memoryDriver) Get(key string) (*RawValue, error) {
	d.mutx.Lock()
	defer d.mutx.Unlock()

	rv, ok := d.values[key]
	if !ok {
		return nil, nil
	}

	if rv.IsExpired() {
		delete(d.values, key)
		return nil, nil
	}

	return rv, nil
}

func (d *memoryDriver) Put(key string, value []byte, expires time.Duration) error {
	d.mutx.Lock()
	defer d.mutx.Unlock()

	d.values[key] = &RawValue{
		Key:     key,
		Value:   value,
		Expires: time.Now().Add(expires),
	}

	return nil
}
