package redis

import (
	"errors"
	"time"
)

var ErrKeyDoesNotExist = errors.New("key does not exist")

type Client interface {
	Set(key, value string, ttl *time.Duration) error
	Get(key string) (string, error)
	Del(key string) error
	Close()
}
