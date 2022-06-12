package mysql

import (
	"github.com/jmoiron/sqlx"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
)

const batchSize = 500

type messageStore struct {
	client Client
}

func (s *messageStore) GetBatch() ([]message.StoredMessage, error) {
	const query = "SELECT `id`, `type`, `topic`, `key`, `body` FROM `message` ORDER BY id ASC LIMIT ?"

	var messagesSqlx []sqlxMessage
	err := s.client.Select(&messagesSqlx, query, batchSize)
	if err != nil {
		return nil, err
	}

	result := make([]message.StoredMessage, 0, len(messagesSqlx))
	for _, sqlxMsg := range messagesSqlx {
		result = append(result, message.StoredMessage{
			ID: sqlxMsg.ID,
			Message: message.Message{
				Type:      sqlxMsg.Type,
				TopicName: sqlxMsg.TopicName,
				Key:       sqlxMsg.Key,
				Body:      sqlxMsg.Body,
			},
		})
	}
	return result, nil
}

func (s *messageStore) Store(msg *message.Message) error {
	const query = "INSERT INTO `message` (`id`, `type`, `topic`, `key`, `body`) VALUES (DEFAULT, :type, :topic, :key, :body)"

	dbMessage := &sqlxMessage{
		Type:      msg.Type,
		TopicName: msg.TopicName,
		Key:       msg.Key,
		Body:      msg.Body,
	}
	_, err := s.client.NamedExec(query, dbMessage)
	return err
}

func (s *messageStore) Delete(ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	query, args, err := sqlx.In("DELETE FROM `message` WHERE id IN (?)", ids)
	if err != nil {
		return err
	}

	_, err = s.client.Exec(query, args...)
	return err
}

func NewMessageStore(client Client) message.Store {
	return &messageStore{client: client}
}

type sqlxMessage struct {
	ID        int    `db:"id"`
	Type      string `db:"type"`
	TopicName string `db:"topic"`
	Key       string `db:"key"`
	Body      []byte `db:"body"`
}
