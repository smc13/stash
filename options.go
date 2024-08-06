package stash

import "log/slog"

type StashOption func(*Stash) error

// WithLogger sets the logger for the Stash instance
func WithLogger(logger *slog.Logger) StashOption {
	return func(s *Stash) error {
		s.logger = logger
		return nil
	}
}

func WithLogErrorKey(key string) StashOption {
	return func(s *Stash) error {
		s.errKey = key
		return nil
	}
}