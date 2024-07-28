package drivers

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilePut(t *testing.T) {
	driver := NewFileDriver(t.TempDir())
	driver.Init()

	err := driver.Put("test", []byte("value"), 5*time.Minute)
	assert.NoError(t, err)
}

func TestFileGet(t *testing.T) {
	driver := NewFileDriver(t.TempDir())
	driver.Init()

	err := driver.Put("test", []byte("value"), 5*time.Minute)
	assert.NoError(t, err)

	value, err := driver.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value"), value)
}

func TestFileForever(t *testing.T) {
	driver := NewFileDriver(t.TempDir())
	driver.Init()

	err := driver.Forever("test", []byte("value"))
	assert.NoError(t, err)

	value, err := driver.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value"), value)
}

func TestFileForget(t *testing.T) {
	dir := t.TempDir()
	driver := NewFileDriver(dir)
	driver.Init()

	var files []os.DirEntry

	assert.NoError(t, driver.Put("test", []byte("value"), 5*time.Minute))
	files, _ = os.ReadDir(dir)
	assert.Len(t, files, 1)

	assert.NoError(t, driver.Forget("test"))

	files, _ = os.ReadDir(dir)
	assert.Empty(t, files)
}

func TestFileFlush(t *testing.T) {
	dir := t.TempDir()
	driver := NewFileDriver(dir)
	driver.Init()

	assert.NoError(t, driver.Put("test", []byte("value"), 5*time.Minute))
	assert.NoError(t, driver.Put("test2", []byte("value"), 5*time.Minute))

	files, _ := os.ReadDir(dir)
	assert.Len(t, files, 2)

	assert.NoError(t, driver.Flush())
	files, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Empty(t, files)
}
