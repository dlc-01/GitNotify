package poller

import (
	"context"
	"log/slog"
	"sync"
	"time"

	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/kafka/producer"
)

type Event struct {
	URL    string
	Source string
	Title  string
	Link   string
}

type Poller interface {
	Source() string
	Supports(url string) bool
	Poll(ctx context.Context, url string, since time.Time) ([]Event, error)
}

type Scheduler struct {
	pollers  []Poller
	producer producer.MultiProducer
	log      *slog.Logger
	interval time.Duration
	mu       sync.RWMutex
	watching map[string]time.Time
}

func NewScheduler(
	producer producer.MultiProducer,
	interval time.Duration,
	log *slog.Logger,
	pollers ...Poller,
) *Scheduler {
	return &Scheduler{
		pollers:  pollers,
		producer: producer,
		log:      log,
		interval: interval,
		watching: make(map[string]time.Time),
	}
}

func (s *Scheduler) Watch(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.watching[url]; !ok {
		s.watching[url] = time.Now()
		s.log.Info("watching url", slog.String("url", url))
	}
}

func (s *Scheduler) Unwatch(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.watching, url)
	s.log.Info("unwatching url", slog.String("url", url))
}

func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.log.Info("scheduler started",
		slog.Duration("interval", s.interval),
	)

	for {
		select {
		case <-ticker.C:
			s.poll(ctx)
		case <-ctx.Done():
			s.log.Info("scheduler stopped")
			return nil
		}
	}
}

func (s *Scheduler) poll(ctx context.Context) {
	s.mu.RLock()
	urls := make(map[string]time.Time, len(s.watching))
	for url, since := range s.watching {
		urls[url] = since
	}
	s.mu.RUnlock()

	for url, since := range urls {
		poller := s.findPoller(url)
		if poller == nil {
			s.log.Warn("no poller for url", slog.String("url", url))
			continue
		}

		events, err := poller.Poll(ctx, url, since)
		if err != nil {
			s.log.Error("poll failed",
				slog.String("url", url),
				slog.String("source", poller.Source()),
				slog.String("err", err.Error()),
			)
			continue
		}

		s.mu.Lock()
		s.watching[url] = time.Now()
		s.mu.Unlock()

		for _, event := range events {
			topic := sourceToTopic(poller.Source())
			if err := s.producer.ProduceTo(ctx, topic, internalkafka.WebhookEventMessage{
				RepoURL:   event.URL,
				EventType: sourceToEventType(poller.Source()),
				Source:    event.Source,
			}); err != nil {
				s.log.Error("produce event",
					slog.String("url", url),
					slog.String("source", poller.Source()),
					slog.String("err", err.Error()),
				)
			}
		}

		if len(events) > 0 {
			s.log.Info("polled events",
				slog.String("url", url),
				slog.String("source", poller.Source()),
				slog.Int("count", len(events)),
			)
		} else {
			s.log.Debug("no new events",
				slog.String("url", url),
				slog.String("source", poller.Source()),
			)
		}
	}
}

func (s *Scheduler) findPoller(url string) Poller {
	for _, p := range s.pollers {
		if p.Supports(url) {
			return p
		}
	}
	return nil
}

func sourceToTopic(source string) internalkafka.Topic {
	switch source {
	case "stackoverflow":
		return internalkafka.TopicEventAnswer
	case "reddit":
		return internalkafka.TopicEventPost
	case "youtube":
		return internalkafka.TopicEventVideo
	default:
		return internalkafka.TopicEventPush
	}
}

func sourceToEventType(source string) string {
	switch source {
	case "stackoverflow":
		return "answer"
	case "reddit":
		return "post"
	case "youtube":
		return "video"
	default:
		return "push"
	}
}
