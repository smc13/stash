package drivers

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// A driver that stores cache values in files
// file names are the SHA1 hash of the key
type fileDriver struct {
	path        string
	permissions os.FileMode
	mutx        sync.Map
}

// Create a new file driver, stored at the given path
func NewFileDriver(path string) Driver {
	return &fileDriver{path: path}
}

// Ensure the folder exists
func (d *fileDriver) Init() error {
	if _, err := os.Stat(d.path); os.IsNotExist(err) {
		return os.MkdirAll(d.path, os.ModePerm)
	}

	return nil
}

func (d *fileDriver) Add(ctx context.Context, raw RawValue) error {
	_, err := d.Get(ctx, raw.Key)

	return err
}

func (d *fileDriver) Flush(_ context.Context) error {
	if _, err := os.Stat(d.path); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(d.path); err != nil {
		return err
	}

	return os.MkdirAll(d.path, os.ModePerm)
}

func (d *fileDriver) Forever(ctx context.Context, raw RawValue) error {
	raw.Expires = time.Unix(9999999999, 0) // the end of Unix time :(
	return d.Put(ctx, raw)
}

func (d *fileDriver) Forget(_ context.Context, key string) (bool, error) {
	unlock := d.lock(key)
	defer unlock()

	err := os.Remove(d.pathForKey(key))
	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func (d *fileDriver) Get(ctx context.Context, key string) (*RawValue, error) {
	unlock := d.lock(key)
	defer unlock()

	return d.getPayload(ctx, key)
}

func (d *fileDriver) Put(_ context.Context, raw RawValue) error {
	unlock := d.lock(raw.Key)
	defer unlock()

	path := d.pathForKey(raw.Key)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "%d", raw.Expires.Unix()); err != nil {
		return err
	}

	_, err = file.Write([]byte(raw.AsString()))
	return err
}

func (d *fileDriver) pathForKey(key string) string {
	hasher := sha1.New()
	hash := hasher.Sum([]byte(key))
	hex := fmt.Sprintf("%x", hash)

	return filepath.Join(d.path, hex)
}

func (d *fileDriver) getPayload(ctx context.Context, key string) (*RawValue, error) {
	path := d.pathForKey(key)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}
	defer file.Close()

	// read the expiration time
	expire := make([]byte, 10)
	if _, err := file.Read(expire); err != nil {
		return nil, err
	}

	expiresAt := time.Unix(int64(binary.BigEndian.Uint64(expire)), 0)

	// check if the value is expired
	if time.Now().After(expiresAt) {
		if _, err := d.Forget(ctx, key); err != nil {
			return nil, err
		}

		return nil, nil
	}

	// read the value
	value, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return RawValueFromString(key, string(value), expiresAt)
}

func (d *fileDriver) lock(key string) func() {
	value, _ := d.mutx.LoadOrStore(key, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()

	return func() {
		d.mutx.Delete(key)
		mtx.Unlock()
	}
}
