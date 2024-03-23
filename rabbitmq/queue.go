package rabbitmq

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type queueService struct {
	ch *amqp.Channel
}

func NewQueueService(url string) (*queueService, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &queueService{
		ch: ch,
	}, nil
}

func (s *queueService) Publish(topic string, message []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	q, err := s.ch.QueueDeclare(
		topic,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	if err := s.ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		},
	); err != nil {
		return err
	}

	return nil
}
