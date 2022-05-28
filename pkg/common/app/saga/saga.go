package saga

import (
	"fmt"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
)

type Operation interface {
	Name() string
	Do() error
	Undo() error
}

type Saga struct {
	name   string
	ops    []Operation
	logger log.Logger
}

func (s *Saga) Execute() (result error) {
	lastCompletedIndex := len(s.ops) - 1
	for i, op := range s.ops {
		err := op.Do()
		if err != nil {
			result = err
			lastCompletedIndex = i - 1
			break
		}
	}

	if lastCompletedIndex == len(s.ops)-1 {
		// saga completed successfully
		return
	}

	// saga failed
	for i := lastCompletedIndex; i >= 0; i-- {
		op := s.ops[i]

		err := op.Undo()
		if err != nil {
			s.handleUndoError(op, err)
		}
	}
	return
}

func (s *Saga) handleUndoError(op Operation, err error) {
	// TODO: must undo operation later
	s.logger.Error(fmt.Errorf("saga \"%s\" rollback failed at \"%s\": %w", s.name, op.Name(), err))
}

func New(name string, ops []Operation, logger log.Logger) *Saga {
	return &Saga{name: name, ops: ops, logger: logger}
}
