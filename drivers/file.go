package drivers

import (
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

func (d *fileDriver) Add(key string, value []byte, expires time.Duration) error {
	_, err := d.Get(key)

	return err
}

func (d *fileDriver) Flush() error {
	if _, err := os.Stat(d.path); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(d.path); err != nil {
		return err
	}

	return os.MkdirAll(d.path, os.ModePerm)
}

func (d *fileDriver) Forever(key string, value []byte) error {
	return d.Put(key, value, 0)
}

func (d *fileDriver) Forget(key string) error {
	unlock := d.lock(key)
	defer unlock()

	return os.Remove(d.pathForKey(key))
}

func (d *fileDriver) Get(key string) (*RawValue, error) {
	unlock := d.lock(key)
	defer unlock()

	return d.getPayload(key)
}

func (d *fileDriver) Put(key string, value []byte, expires time.Duration) error {
	unlock := d.lock(key)
	defer unlock()

	path := d.pathForKey(key)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "%d", d.expires(expires)); err != nil {
		return err
	}

	_, err = file.Write(value)
	return err
}

func (d *fileDriver) pathForKey(key string) string {
	hasher := sha1.New()
	hash := hasher.Sum([]byte(key))
	hex := fmt.Sprintf("%x", hash)

	return filepath.Join(d.path, hex)
}

// Calculate the expiration time for a cache value as a Unix timestamp
func (d *fileDriver) expires(duration time.Duration) int64 {
	t := time.Now().Add(duration).Unix()
	if duration == 0 || t > 9999999999 {
		return 9999999999
	}

	return t
}

func (d *fileDriver) getPayload(key string) (*RawValue, error) {
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
		if err := d.Forget(key); err != nil {
			return nil, err
		}

		return nil, nil
	}

	// read the value
	value, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return &RawValue{key, value, expiresAt}, nil
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
