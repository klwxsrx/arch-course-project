package redis

import (
	"context"
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

func (c *connection) Set(key, value string, ttl time.Duration) error {
	return c.client.SetEX(context.Background(), key, value, ttl).Err()
}

func (c *connection) Get(key string) (string, error) {
	return c.client.Get(context.Background(), key).Result()
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
