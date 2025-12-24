package analytics

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

type Event struct {
	Type       string      `json:"type"`
	GameID     string      `json:"gameId"`
	Payload    interface{} `json:"payload"`
	OccurredAt time.Time   `json:"occurredAt"`
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			RequiredAcks: kafka.RequireAll,
			Balancer:     &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Emit(ctx context.Context, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Value: payload})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
