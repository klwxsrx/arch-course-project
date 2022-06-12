package mysql

import (
	"fmt"

	"github.com/klwxsrx/arch-course-project/pkg/common/app/persistence"
)

type synchronization struct {
	client TransactionalClient
}

func (s *synchronization) CriticalSection(name string, f func()) error {
	l := NewLock(s.client, name)
	err := l.Get()
	if err != nil {
		return fmt.Errorf("can't create db lock %s: %w", name, err)
	}
	defer l.Release()

	f()
	return nil
}

func NewSynchronization(client TransactionalClient) persistence.Synchronization {
	return &synchronization{client}
}
