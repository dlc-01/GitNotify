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
