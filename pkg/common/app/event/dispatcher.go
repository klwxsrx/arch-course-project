package event

import (
	"fmt"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
)

type Event struct {
	Type      string
	TopicName string
	Key       string
	Body      []byte
}

type Dispatcher interface {
	Dispatch(msg *Event) error
}

type dispatcher struct {
	messageStore message.Store
}

func (d *dispatcher) Dispatch(msg *Event) error {
	err := d.messageStore.Store(&message.Message{
		Type:      msg.Type,
		TopicName: msg.TopicName,
		Key:       msg.Key,
		Body:      msg.Body,
	})
	if err != nil {
		return fmt.Errorf("failed to dispatch message: %w", err)
	}
	return nil
}

func NewDispatcher(messageStore message.Store) Dispatcher {
	return &dispatcher{messageStore: messageStore}
}
