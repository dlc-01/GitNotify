package parser

import "github.com/dlc-01/GitNotify/internal/domain"

type Event struct {
	RepoURL   string
	EventType domain.EventType
	Source    string
}

type Parser interface {
	Source() string
	Parse(eventType string, payload []byte) (*Event, error)
}
