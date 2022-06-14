package pulsar

import (
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
)

type subscription struct{ Topic, Type string }

type consumer struct {
	c        pulsar.Consumer
	stopChan chan struct{}
}

type MessageSubscriber struct {
	subscriberName string
	conn           Connection
	consumers      map[string]consumer
	logger         log.Logger
	handlers       map[subscription]message.Handler
}

func (s *MessageSubscriber) subscribe(handler message.Handler) {
	s.handlers[subscription{
		Topic: getTopicFullName(handler.TopicName()),
		Type:  handler.Type(),
	}] = handler
}

func (s *MessageSubscriber) run() error {
	for subscription := range s.handlers {
		err := s.runConsumerIfDoesntExist(subscription.Topic)
		if err != nil {
			return fmt.Errorf("failed to run consumer for topic %s: %w", subscription.Topic, err)
		}
	}
	return nil
}

func (s *MessageSubscriber) close() {
	for _, c := range s.consumers {
		c.stopChan <- struct{}{}
	}
}

func (s *MessageSubscriber) runConsumerIfDoesntExist(topic string) error {
	if _, ok := s.consumers[topic]; ok {
		return nil
	}

	c, err := s.conn.Subscribe(&ConsumerConfig{
		Topic:            topic,
		SubscriptionName: s.subscriberName,
	})
	if err != nil {
		return err
	}

	stopChan := make(chan struct{})
	s.consumers[topic] = consumer{
		c:        c,
		stopChan: stopChan,
	}
	go func() {
		for {
			select {
			case msg, ok := <-c.Chan():
				if !ok {
					return
				}

				s.processMessage(&msg)
			case <-stopChan:
				return
			}
		}
	}()

	return nil
}

func (s *MessageSubscriber) processMessage(msg *pulsar.ConsumerMessage) {
	typ, ok := msg.Properties()[propertyMessageType]
	if !ok {
		msg.Consumer.Ack(msg)
		return
	}

	handler, ok := s.handlers[subscription{Topic: msg.Topic(), Type: typ}]
	if !ok {
		return
	}

	err := handler.Handle(&message.Message{
		Type:      typ,
		TopicName: msg.Topic(),
		Key:       msg.Key(),
		Body:      msg.Payload(),
	})
	if err != nil {
		s.logger.WithError(err).Error(fmt.Sprintf("failed to handle message %s", msg.Payload()))
		msg.Consumer.Nack(msg)
		return
	}

	s.logger.Info(fmt.Sprintf("handled message %s: with key %s", typ, msg.Key()))
	msg.Consumer.Ack(msg)
}

func NewMessageSubscriber(
	subscriberName string,
	handlers []message.Handler,
	conn Connection,
	logger log.Logger,
) (closer func(), err error) {
	s := &MessageSubscriber{
		subscriberName: subscriberName,
		conn:           conn,
		consumers:      make(map[string]consumer),
		logger:         logger,
		handlers:       make(map[subscription]message.Handler),
	}

	for _, handler := range handlers {
		s.subscribe(handler)
	}

	err = s.run()
	if err != nil {
		return func() {}, err
	}
	return s.close, nil
}
