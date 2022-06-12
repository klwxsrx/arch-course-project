package pulsar

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
)

const propertyMessageType = "type"

type messageSender struct {
	conn      Connection
	producers map[string]pulsar.Producer
}

func (s *messageSender) Send(msg *message.Message) error {
	producer, err := s.getProducerForTopic(msg.TopicName)
	if err != nil {
		return fmt.Errorf("failed to create producer for topic %s: %w", msg.TopicName, err)
	}

	_, err = producer.Send(context.Background(), &pulsar.ProducerMessage{
		Payload:    msg.Body,
		Key:        msg.Key,
		Properties: map[string]string{propertyMessageType: msg.Type},
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (s *messageSender) getProducerForTopic(topic string) (pulsar.Producer, error) {
	if p, ok := s.producers[topic]; ok {
		return p, nil
	}

	p, err := s.conn.CreateProducer(&ProducerConfig{Topic: topic})
	if err != nil {
		return nil, err
	}

	s.producers[topic] = p
	return p, nil
}

func NewMessageSender(conn Connection) message.Sender {
	return &messageSender{conn: conn, producers: make(map[string]pulsar.Producer)}
}
