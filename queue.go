package blog

type QueueService interface {
	Publish(topic string, message []byte) error
}
