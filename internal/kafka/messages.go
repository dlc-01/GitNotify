package kafka

type SubscriptionCreatedMessage struct {
	ChatID  int64  `json:"chat_id"`
	RepoURL string `json:"repo_url"`
}

type SubscriptionDeletedMessage struct {
	ChatID  int64  `json:"chat_id"`
	RepoURL string `json:"repo_url"`
}

type SubscriptionMutedMessage struct {
	ChatID  int64  `json:"chat_id"`
	RepoURL string `json:"repo_url"`
	Event   string `json:"event"`
}

type WebhookEventMessage struct {
	RepoURL   string `json:"repo_url"`
	EventType string `json:"event_type"`
	Source    string `json:"source"`
}

type SubscriptionUnmutedMessage struct {
	ChatID  int64  `json:"chat_id"`
	RepoURL string `json:"repo_url"`
	Event   string `json:"event"`
}
