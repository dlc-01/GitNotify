package producer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"

	kafkapkg "github.com/dlc-01/GitNotify/internal/kafka"
)

type Producer interface {
	Produce(ctx context.Context, msg any) error
	Close() error
}

type producer struct {
	writer *kafka.Writer
}

func New(brokers []string, topic kafkapkg.Topic) Producer {
	return &producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic.String(),
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: 10 * time.Second,
			RequiredAcks: kafka.RequireOne,
		},
	}
}

func (p *producer) Produce(ctx context.Context, msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return wrap("Produce", p.writer.Topic, ErrMarshal)
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{Value: data}); err != nil {
		return wrap("Produce", p.writer.Topic, ErrProduce)
	}
	return nil
}

func (p *producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return wrap("Close", p.writer.Topic, err)
	}
	return nil
}
