package redis

import "time"

type Client interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, error)
	Del(key string) error
	Close()
}
