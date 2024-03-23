package blog

import "context"

type Event struct {
	UserID    string `json:"user_id"`
	IP        string `json:"ip"`
	Country   string `json:"country"`
	UserAgent string `json:"user_agent"`
	Browser   string `json:"browser"`
	OS        string `json:"os"`
	Referer   string `json:"referer"`
	URL       string `json:"url"`
	Time      string `json:"time"`
}

type EventService interface {
	SendMessage(topic, key string, value []byte) error
	Consume(ctx context.Context, topic string) (<-chan *Event, error)
}
