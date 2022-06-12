package message

type Message struct {
	Type      string
	TopicName string
	Key       string
	Body      []byte
}

type StoredMessage struct {
	ID int
	Message
}

type Store interface {
	GetBatch() ([]StoredMessage, error)
	Store(msg *Message) error
	Delete(ids []int) error
}
