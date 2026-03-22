package kafka

type Topic string

const (
	TopicSubscriptionCreated Topic = "subscriptions.created"
	TopicSubscriptionDeleted Topic = "subscriptions.deleted"
	TopicSubscriptionMuted   Topic = "subscriptions.muted"
	TopicEventPush           Topic = "events.push"
	TopicEventPR             Topic = "events.pr"
	TopicEventIssue          Topic = "events.issue"
	TopicEventPipeline       Topic = "events.pipeline"
)

func (t Topic) String() string {
	return string(t)
}
