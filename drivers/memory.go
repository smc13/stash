package drivers

import (
	"context"
	"sync"
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

func (d *memoryDriver) Add(_ context.Context, raw RawValue) error {
	d.mutx.Lock()
	defer d.mutx.Unlock()

	val, found := d.values[raw.Key]
	if found && !val.IsExpired() {
		return nil
	}

	d.values[raw.Key] = &raw
	return nil
}

func (d *memoryDriver) Flush(_ context.Context) error {
	d.mutx.Lock()
	defer d.mutx.Unlock()

	d.values = make(map[string]*RawValue)
	return nil
}

func (d *memoryDriver) Forever(ctx context.Context, raw RawValue) error {
	return d.Put(ctx, raw)
}

func (d *memoryDriver) Forget(_ context.Context, key string) (bool, error) {
	d.mutx.Lock()
	defer d.mutx.Unlock()

	if _, ok := d.values[key]; !ok {
		return false, nil
	}

	delete(d.values, key)
	return true, nil
}

func (d *memoryDriver) Get(_ context.Context, key string) (*RawValue, error) {
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

func (d *memoryDriver) Put(_ context.Context, raw RawValue) error {
	d.mutx.Lock()
	defer d.mutx.Unlock()

	d.values[raw.Key] = &raw
	return nil
}
