package stash

import "context"

func Get[T any](ctx context.Context, stash *Stash, key string) (T, error) {
	var value T

	err := stash.Get(ctx, key, &value)
	return value, err
}
