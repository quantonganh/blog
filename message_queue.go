package blog

type MessageQueueService interface {
	Publish(topic string, message []byte) error
}
