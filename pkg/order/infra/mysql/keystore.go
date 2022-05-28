package mysql

import (
	"github.com/klwxsrx/arch-course-project/pkg/common/app/idempotence"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
)

type idempotenceKeyStore struct {
	client mysql.Client
}

func (s *idempotenceKeyStore) StoreUnique(key string) error {
	result, err := s.client.Exec("INSERT INTO idk (`key`) VALUES (?) ON DUPLICATE KEY UPDATE `key`=`key`", key)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return idempotence.ErrKeyAlreadyExists
	}
	return nil
}

func NewIdempotenceKeyStore(client mysql.Client) idempotence.KeyStore {
	return &idempotenceKeyStore{client: client}
}
