package callback

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dlc-01/GitNotify/internal/domain"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/repository"
)

func TestMuteHandler_Execute_Success(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	prod := &mockMultiProducer{}
	log := newTestLogger()

	repo.On("MuteEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	prod.On("ProduceTo", mock.Anything, internalkafka.TopicSubscriptionMuted, mock.Anything).Return(nil)

	h := NewMuteHandler(repo, sender, log, prod)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go:push")

	require.NotEmpty(t, sender.edited)
	assert.Contains(t, sender.edited[0], "Muted")
	assert.Contains(t, sender.edited[0], "push")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}

func TestMuteHandler_Execute_InvalidArgs(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	prod := &mockMultiProducer{}
	log := newTestLogger()

	h := NewMuteHandler(repo, sender, log, prod)
	h.Execute(context.Background(), 123, 1, "invaliddatawithoutseparator")

	assert.Contains(t, sender.answered[0], "Invalid callback data")
	assert.Empty(t, sender.edited)
	repo.AssertNotCalled(t, "MuteEvent")
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestMuteHandler_Execute_InvalidEvent(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	prod := &mockMultiProducer{}
	log := newTestLogger()

	h := NewMuteHandler(repo, sender, log, prod)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go:invalidevent")

	assert.Contains(t, sender.answered[0], "Invalid event type")
	assert.Empty(t, sender.edited)
	repo.AssertNotCalled(t, "MuteEvent")
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestMuteHandler_Execute_NotFound(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	prod := &mockMultiProducer{}
	log := newTestLogger()

	repo.On("MuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(repository.Wrap("MuteEvent", repository.ErrNotFound))

	h := NewMuteHandler(repo, sender, log, prod)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go:push")

	assert.Contains(t, sender.answered[0], "Subscription not found")
	assert.Empty(t, sender.edited)
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestMuteHandler_Execute_RepoError(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	prod := &mockMultiProducer{}
	log := newTestLogger()

	repo.On("MuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(repository.Wrap("MuteEvent", repository.ErrInvalidInput))

	h := NewMuteHandler(repo, sender, log, prod)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go:push")

	assert.Contains(t, sender.answered[0], "Internal error")
	assert.Empty(t, sender.edited)
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestMuteHandler_Execute_ProduceError(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	prod := &mockMultiProducer{}
	log := newTestLogger()

	repo.On("MuteEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	prod.On("ProduceTo", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

	h := NewMuteHandler(repo, sender, log, prod)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go:push")

	assert.Contains(t, sender.edited[0], "Muted")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}

func TestMuteHandler_Action(t *testing.T) {
	h := &MuteHandler{}
	assert.Equal(t, "mute", h.Action())
}
