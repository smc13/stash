package stash

type Stashable interface {
	ToStash() ([]byte, error)
	FromStash([]byte) error
}
