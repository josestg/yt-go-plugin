package cache

import (
	"context"
	"log/slog"
	"time"
)

// consterror is a custom error type used to represent specific errors in the cache implementation.
// It is derived from the int type to allow it to be used as a constant, ensuring immutability across packages.
type consterror int

// Possible errors returned by the cache implementation.
const (
	ErrNotFound consterror = iota
	ErrExpired
)

// _text maps consterror values to their corresponding error messages.
var _text = map[consterror]string{
	ErrNotFound: "cache: key not found",
	ErrExpired:  "cache: key expired",
}

// Error implements the error interface.
func (e consterror) Error() string {
	txt, ok := _text[e]
	if !ok {
		return "cache: unknown error"
	}
	return txt
}

// Cache defines the interface for a cache implementation.
type Cache interface {
	// Set stores a key-value pair in the cache with a specified expiration time.
	Set(ctx context.Context, key, val string, exp time.Duration) error

	// Get retrieves a value from the cache by its key.
	// Returns ErrNotFound if the key is not found.
	// Returns ErrExpired if the key has expired.
	Get(ctx context.Context, key string) (string, error)
}

// Factory defines the function signature for creating a cache implementation.
type Factory func(log *slog.Logger) (Cache, error)

// nopCache is a no-operation cache implementation.
type nopCache int

// NopCache a singleton cache instance, which does nothing.
const NopCache nopCache = 0

// Ensure that NopCache implements the Cache interface.
var _ Cache = NopCache

// Set is a no-op and always returns nil.
func (nopCache) Set(context.Context, string, string, time.Duration) error { return nil }

// Get always returns ErrNotFound, indicating that the key does not exist in the cache.
func (nopCache) Get(context.Context, string) (string, error) { return "", ErrNotFound }
