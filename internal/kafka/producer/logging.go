package producer

import (
	"context"
	"log/slog"
	"time"
)

type loggingProducer struct {
	producer Producer
	topic    string
	log      *slog.Logger
}

func NewLogging(p Producer, topic string, log *slog.Logger) Producer {
	return &loggingProducer{producer: p, topic: topic, log: log}
}

func (p *loggingProducer) Produce(ctx context.Context, msg any) error {
	start := time.Now()
	err := p.producer.Produce(ctx, msg)
	p.log.Debug("produce",
		slog.String("topic", p.topic),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	return err
}

func (p *loggingProducer) Close() error {
	return p.producer.Close()
}
