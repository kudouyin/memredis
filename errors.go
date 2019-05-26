package memredis

import "errors"

var (
	ErrKeyNotFound = errors.New("Key not found in cache")
)
