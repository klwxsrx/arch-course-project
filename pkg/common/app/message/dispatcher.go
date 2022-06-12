package message

import (
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/persistence"
	"sync"
)

type Sender interface {
	Send(msg *Message) error
}

type Dispatcher interface {
	Dispatch()
	Close()
}

type dispatcher struct {
	store  Store
	sender Sender
	sync   persistence.Synchronization
	logger log.Logger

	dispatchChan chan struct{}
	stopChan     chan struct{}
	onceCloser   *sync.Once
}

func (d *dispatcher) Dispatch() {
	select {
	case d.dispatchChan <- struct{}{}:
	default:
	}
}

func (d *dispatcher) Close() {
	d.onceCloser.Do(func() {
		d.stopChan <- struct{}{}
	})
}

func (d *dispatcher) run() {
	for {
		select {
		case <-d.dispatchChan:
			d.processMessages()
		case <-d.stopChan:
			return
		}
	}
}

func (d *dispatcher) processMessages() {
	processBatch := func() (batchProcessed bool) {
		msgs, err := d.store.GetBatch()
		if err != nil {
			d.logger.WithError(err).Error("failed to get messages for send")
			return false
		}
		if len(msgs) == 0 {
			return false
		}

		for _, msg := range msgs {
			err := d.sender.Send(&msg.Message)
			if err != nil {
				d.logger.WithError(err).Error("failed to send message")
				return false
			}
			d.logger.With(log.Fields{"id": msg.ID, "type": msg.Type, "topic": msg.TopicName}).Info("message sent")

			err = d.store.Delete([]int{msg.ID})
			if err != nil {
				d.logger.WithError(err).Error("failed to delete sent messages")
				return false
			}
		}
		return true
	}

	err := d.sync.CriticalSection("process_message_dispatch", func() {
		for processBatch() {
		}
	})
	if err != nil {
		d.logger.WithError(err).Error("failed to process messages")
	}
}

func NewDispatcher(store Store, sender Sender, synchro persistence.Synchronization, logger log.Logger) Dispatcher {
	d := &dispatcher{
		store:        store,
		sender:       sender,
		sync:         synchro,
		logger:       logger,
		dispatchChan: make(chan struct{}, 1),
		stopChan:     make(chan struct{}),
		onceCloser:   &sync.Once{},
	}
	go d.run()
	return d
}
