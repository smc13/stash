package stash

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
	prefix      string
	mutx        sync.Map
}

// Create a new file driver, stored at the given path
func NewFileDriver(path string) *fileDriver {
	return &fileDriver{path: path}
}

// Ensure the folder exists
func (d *fileDriver) Init() error {
	if _, err := os.Stat(d.path); os.IsNotExist(err) {
		return os.MkdirAll(d.path, os.ModePerm)
	}

	return nil
}

func (d *fileDriver) Prefix(prefix string) Driver {
	d.prefix = prefix
	return d
}

func (d *fileDriver) Add(ctx context.Context, raw CacheItem) error {
	_, err := d.Get(ctx, raw.Key)

	return err
}

func (d *fileDriver) Flush(_ context.Context) error {
	if _, err := os.Stat(d.path); os.IsNotExist(err) {
		return nil
	}

	path := filepath.Join(d.path, d.prefix)
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return os.MkdirAll(d.path, os.ModePerm)
}

func (d *fileDriver) Forever(ctx context.Context, raw CacheItem) error {
	raw.Expires = time.Time{}
	return d.Put(ctx, raw)
}

func (d *fileDriver) Forget(_ context.Context, key string) (bool, error) {
	unlock := d.lock(key)
	defer unlock()

	path, err := d.pathForKey(key, false)
	if err != nil {
		return false, err
	}

	err = os.Remove(path)
	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func (d *fileDriver) Get(ctx context.Context, key string) (*CacheItem, error) {
	unlock := d.lock(key)
	defer unlock()

	return d.getPayload(ctx, key)
}

func (d *fileDriver) Put(_ context.Context, raw CacheItem) error {
	unlock := d.lock(raw.Key)
	defer unlock()

	path, err := d.pathForKey(raw.Key, true)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	unix := raw.Expires.Unix()
	if raw.Expires.IsZero() {
		unix = 9999999999
	}

	if _, err := fmt.Fprintf(file, "%d", unix); err != nil {
		return err
	}

	_, err = file.Write([]byte(raw.Value))
	return err
}

func (d *fileDriver) pathForKey(key string, create bool) (string, error) {
	hasher := sha1.New()
	hash := hasher.Sum([]byte(key))
	hex := fmt.Sprintf("%x", hash)

	path := filepath.Join(d.path, d.prefix, hex)

	if !create {
		return path, nil
	}

	return path, os.MkdirAll(filepath.Dir(path), os.ModePerm)
}

func (d *fileDriver) getPayload(ctx context.Context, key string) (*CacheItem, error) {
	path, err := d.pathForKey(key, false)
	if err != nil {
		return nil, err
	}

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

	if len(expire) != 10 {
		return nil, fmt.Errorf("invalid expiration time")
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

	return &CacheItem{Key: key, Value: string(value), Expires: expiresAt}, nil
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
