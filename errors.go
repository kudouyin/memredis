package memredis

import "errors"

var (
	ErrKeyNotFound = errors.New("Key not found in cache")
	ErrCreateEventBase = errors.New("create event base error")
	ErrAddEvent = errors.New("add event error")
	ErrWaitEvent = errors.New("wait event error")
)
