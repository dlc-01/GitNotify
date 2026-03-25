package kafka

type Topic string

const (
	TopicSubscriptionCreated Topic = "subscriptions.created"
	TopicSubscriptionDeleted Topic = "subscriptions.deleted"
	TopicSubscriptionMuted   Topic = "subscriptions.muted"
	TopicSubscriptionUnmuted Topic = "subscriptions.unmuted"

	TopicEventPush     Topic = "events.push"
	TopicEventPR       Topic = "events.pr"
	TopicEventIssue    Topic = "events.issue"
	TopicEventPipeline Topic = "events.pipeline"
	TopicEventAnswer   Topic = "events.answer"
	TopicEventPost     Topic = "events.post"
	TopicEventVideo    Topic = "events.video"
)

func (t Topic) String() string {
	return string(t)
}
