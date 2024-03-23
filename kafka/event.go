package kafka

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/quantonganh/blog"
	"github.com/rs/zerolog/log"
)

type eventService struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
}

func NewEventService(brokerAddr string) (*eventService, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{brokerAddr}, config)
	if err != nil {
		return nil, err
	}

	consumer, err := sarama.NewConsumer([]string{brokerAddr}, config)
	if err != nil {
		return nil, err
	}

	return &eventService{
		producer: producer,
		consumer: consumer,
	}, nil
}

func (es *eventService) SendMessage(topic, key string, value []byte) error {
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	_, _, err := es.producer.SendMessage(message)
	if err != nil {
		return err
	}

	return nil
}

func (es *eventService) Consume(ctx context.Context, topic string) (<-chan *blog.Event, error) {
	partition := int32(0)
	offset := int64(sarama.OffsetNewest)
	pc, err := es.consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return nil, err
	}

	c := make(chan *blog.Event)
	go func() {
		defer close(c)

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-pc.Messages():
				var e *blog.Event
				if err := json.Unmarshal(msg.Value, &e); err != nil {
					log.Error().Err(err).Msg("failed to unmarshal message value")
					return
				}
				e.UserID = string(msg.Key)

				c <- e
			}
		}
	}()

	return c, nil
}
