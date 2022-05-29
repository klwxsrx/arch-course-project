package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/go-redis/redis/v8"
	"time"
)

const connTimeout = 30 * time.Second

type Config struct {
	Address  string
	Password string
}

type connection struct {
	client *redis.Client
}

func (c *connection) Set(key, value string, ttl *time.Duration) error {
	if ttl == nil {
		return c.client.Set(context.Background(), key, value, 0).Err()
	}
	return c.client.Set(context.Background(), key, value, *ttl).Err()
}

func (c *connection) Get(key string) (string, error) {
	val, err := c.client.Get(context.Background(), key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrKeyDoesNotExist
	}
	return val, err
}

func (c *connection) Del(key string) error {
	return c.client.Del(context.Background(), key).Err()
}

func (c *connection) Close() {
	_ = c.client.Close()
}

func newOpenConnectionBackoff(connTimeout time.Duration) *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = connTimeout
	return b
}

func NewClient(config *Config) (Client, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
	})

	err := backoff.Retry(func() error {
		return cli.Ping(context.Background()).Err()
	}, newOpenConnectionBackoff(connTimeout))
	if err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("failed to open redis connection: %w", err)
	}

	return &connection{client: cli}, nil
}
