package main

import (
	"fmt"
	"os"
)

type config struct {
	RedisAddress      string
	RedisPassword     string
	OrderServiceURL   string
	CatalogServiceURL string
}

func parseEnvString(key string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	str, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("undefined environment variable %s", key)
	}
	return str, nil
}

func parseConfig() (*config, error) {
	var err error
	redisAddress, err := parseEnvString("REDIS_ADDRESS", err)
	redisPassword, err := parseEnvString("REDIS_PASSWORD", err)
	orderServiceURL, err := parseEnvString("ORDER_SERVICE_URL", err)
	catalogServiceURL, err := parseEnvString("CATALOG_SERVICE_URL", err)

	if err != nil {
		return nil, err
	}

	return &config{
		redisAddress,
		redisPassword,
		orderServiceURL,
		catalogServiceURL,
	}, nil
}
