package cache

import "errors"

var (
	ErrNotFound     = errors.New("cache: key not found")
	ErrItemTooLarge = errors.New("cache: item exceeds max bytes limit")
)
