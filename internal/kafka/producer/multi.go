package producer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	kafkapkg "github.com/dlc-01/GitNotify/internal/kafka"
)

type MultiProducer interface {
	ProduceTo(ctx context.Context, topic kafkapkg.Topic, msg any) error
	Close() error
}

type multiProducer struct {
	producers map[kafkapkg.Topic]Producer
}

func NewMulti(brokers []string, topics ...kafkapkg.Topic) MultiProducer {
	producers := make(map[kafkapkg.Topic]Producer)
	for _, topic := range topics {
		producers[topic] = New(brokers, topic)
	}
	return &multiProducer{producers: producers}
}

func (m *multiProducer) ProduceTo(ctx context.Context, topic kafkapkg.Topic, msg any) error {
	p, ok := m.producers[topic]
	if !ok {
		return fmt.Errorf("no producer for topic %s", topic)
	}
	return p.Produce(ctx, msg)
}

func (m *multiProducer) Close() error {
	for _, p := range m.producers {
		if err := p.Close(); err != nil {
			return err
		}
	}
	return nil
}

type loggingMultiProducer struct {
	producer MultiProducer
	log      *slog.Logger
}

func NewLoggingMulti(p MultiProducer, log *slog.Logger) MultiProducer {
	return &loggingMultiProducer{producer: p, log: log}
}

func (p *loggingMultiProducer) ProduceTo(ctx context.Context, topic kafkapkg.Topic, msg any) error {
	start := time.Now()
	err := p.producer.ProduceTo(ctx, topic, msg)
	p.log.Debug("produce",
		slog.String("topic", topic.String()),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	return err
}

func (p *loggingMultiProducer) Close() error {
	return p.producer.Close()
}
