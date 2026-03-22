package domain

import "time"

type EventType string

const (
	EventPush     EventType = "push"
	EventPR       EventType = "pr"
	EventIssue    EventType = "issue"
	EventPipeline EventType = "pipeline"
)

var AllEventTypes = []EventType{EventPush, EventPR, EventIssue, EventPipeline}

func (e EventType) Valid() bool {
	for _, t := range AllEventTypes {
		if e == t {
			return true
		}
	}
	return false
}

type User struct {
	UserID    int64
	Username  string
	CreatedAt time.Time
}

type ChatType string

const (
	ChatPrivate    ChatType = "private"
	ChatGroup      ChatType = "group"
	ChatSupergroup ChatType = "supergroup"
)

type Chat struct {
	ChatID    int64
	ChatType  ChatType
	CreatedAt time.Time
}

type Subscription struct {
	ID          int64
	ChatID      int64
	RepoURL     string
	MutedEvents []EventType
	CreatedAt   time.Time
}

func (s *Subscription) IsEventMuted(e EventType) bool {
	for _, m := range s.MutedEvents {
		if m == e {
			return true
		}
	}
	return false
}
