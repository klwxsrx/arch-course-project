package message

type Handler interface {
	TopicName() string
	Type() string
	Handle(msg *Message) error
}
