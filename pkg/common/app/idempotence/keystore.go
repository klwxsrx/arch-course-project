package idempotence

import (
	"errors"
)

var ErrKeyAlreadyExists = errors.New("idempotence key is already exist")

type KeyStore interface {
	StoreUnique(key string) error
}
