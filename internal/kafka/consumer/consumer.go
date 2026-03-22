package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"

	kafkapkg "github.com/dlc-01/GitNotify/internal/kafka"
)

type Handler func(ctx context.Context, msg []byte) error

type Consumer struct {
	brokers []string
	groupID string
	log     *slog.Logger
	readers map[kafkapkg.Topic]*kafka.Reader
}

func New(brokers []string, groupID string, log *slog.Logger) *Consumer {
	return &Consumer{
		brokers: brokers,
		groupID: groupID,
		log:     log,
		readers: make(map[kafkapkg.Topic]*kafka.Reader),
	}
}

func (c *Consumer) Subscribe(topic kafkapkg.Topic, handler Handler) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.brokers,
		Topic:   topic.String(),
		GroupID: c.groupID,
	})
	c.readers[topic] = r

	go c.consume(topic, r, handler)
}

func (c *Consumer) consume(topic kafkapkg.Topic, r *kafka.Reader, handler Handler) {
	for {
		msg, err := r.ReadMessage(context.Background())
		if err != nil {
			c.log.Error("read message",
				"topic", topic.String(),
				"err", wrap("consume", topic.String(), ErrConsume),
			)
			continue
		}

		if err := handler(context.Background(), msg.Value); err != nil {
			c.log.Error("handle message",
				"topic", topic.String(),
				"err", err,
			)
		}
	}
}

func (c *Consumer) Close() error {
	for topic, r := range c.readers {
		if err := r.Close(); err != nil {
			return wrap("Close", topic.String(), err)
		}
	}
	return nil
}

func Unmarshal[T any](data []byte) (*T, error) {
	var msg T
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, &Error{Op: "Unmarshal", Err: ErrUnmarshal}
	}
	return &msg, nil
}
