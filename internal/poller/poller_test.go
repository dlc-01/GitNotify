package poller

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func TestStackOverflowPoller_Source(t *testing.T) {
	p := NewStackOverflowPoller()
	assert.Equal(t, "stackoverflow", p.Source())
}

func TestStackOverflowPoller_Supports(t *testing.T) {
	p := NewStackOverflowPoller()
	assert.True(t, p.Supports("https://stackoverflow.com/questions/tagged/golang"))
	assert.False(t, p.Supports("https://github.com/golang/go"))
}

func TestStackOverflowPoller_Poll_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": [
			{"question_id": 1, "title": "How to use goroutines?", "link": "https://stackoverflow.com/q/1", "creation_date": 9999999999},
			{"question_id": 2, "title": "What is a channel?", "link": "https://stackoverflow.com/q/2", "creation_date": 9999999999}
		]}`))
	}))
	defer srv.Close()

	p := NewStackOverflowPoller()
	p.client = srv.Client()
	p.apiURL = srv.URL

	events, err := p.Poll(context.Background(), "https://stackoverflow.com/questions/tagged/golang", time.Now().Add(-time.Hour))
	require.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, "stackoverflow", events[0].Source)
	assert.Equal(t, "How to use goroutines?", events[0].Title)
}

func TestStackOverflowPoller_Poll_FiltersOldQuestions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": [
			{"question_id": 1, "title": "Old question", "link": "https://stackoverflow.com/q/1", "creation_date": 0}
		]}`))
	}))
	defer srv.Close()

	p := NewStackOverflowPoller()
	p.client = srv.Client()
	p.apiURL = srv.URL

	events, err := p.Poll(context.Background(), "https://stackoverflow.com/questions/tagged/golang", time.Now())
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestStackOverflowPoller_Poll_InvalidURL(t *testing.T) {
	p := NewStackOverflowPoller()
	_, err := p.Poll(context.Background(), "https://stackoverflow.com/questions/tagged/", time.Now())
	require.Error(t, err)

	var pollerErr *Error
	require.ErrorAs(t, err, &pollerErr)
	assert.ErrorIs(t, pollerErr, ErrInvalidURL)
}

func TestStackOverflowPoller_Poll_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := NewStackOverflowPoller()
	p.client = srv.Client()
	p.apiURL = srv.URL

	_, err := p.Poll(context.Background(), "https://stackoverflow.com/questions/tagged/golang", time.Now())
	require.Error(t, err)

	var pollerErr *Error
	require.ErrorAs(t, err, &pollerErr)
	assert.ErrorIs(t, pollerErr, ErrFetch)
}

func TestStackOverflowPoller_Poll_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer srv.Close()

	p := NewStackOverflowPoller()
	p.client = srv.Client()
	p.apiURL = srv.URL

	_, err := p.Poll(context.Background(), "https://stackoverflow.com/questions/tagged/golang", time.Now())
	require.Error(t, err)

	var pollerErr *Error
	require.ErrorAs(t, err, &pollerErr)
	assert.ErrorIs(t, pollerErr, ErrParse)
}

func TestRedditPoller_Source(t *testing.T) {
	p := NewRedditPoller()
	assert.Equal(t, "reddit", p.Source())
}

func TestRedditPoller_Supports(t *testing.T) {
	p := NewRedditPoller()
	assert.True(t, p.Supports("https://reddit.com/r/golang"))
	assert.True(t, p.Supports("https://www.reddit.com/r/golang"))
	assert.False(t, p.Supports("https://github.com/golang/go"))
}

func TestRedditPoller_Poll_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom">
			<entry>
				<title>New Go release</title>
				<link href="https://reddit.com/r/golang/comments/1"/>
				<updated>2099-01-01T00:00:00Z</updated>
			</entry>
			<entry>
				<title>Go generics explained</title>
				<link href="https://reddit.com/r/golang/comments/2"/>
				<updated>2099-01-01T00:00:00Z</updated>
			</entry>
		</feed>`))
	}))
	defer srv.Close()

	p := &RedditPoller{client: srv.Client()}
	events, err := p.Poll(context.Background(), srv.URL+"/r/golang", time.Now().Add(-time.Hour))
	require.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, "reddit", events[0].Source)
	assert.Equal(t, "New Go release", events[0].Title)
}

func TestRedditPoller_Poll_FiltersOldPosts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom">
			<entry>
				<title>Old post</title>
				<link href="https://reddit.com/r/golang/comments/1"/>
				<updated>2000-01-01T00:00:00Z</updated>
			</entry>
		</feed>`))
	}))
	defer srv.Close()

	p := &RedditPoller{client: srv.Client()}
	events, err := p.Poll(context.Background(), srv.URL+"/r/golang", time.Now())
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestRedditPoller_Poll_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	p := &RedditPoller{client: srv.Client()}
	_, err := p.Poll(context.Background(), srv.URL+"/r/golang", time.Now())
	require.Error(t, err)

	var pollerErr *Error
	require.ErrorAs(t, err, &pollerErr)
	assert.ErrorIs(t, pollerErr, ErrFetch)
}

func TestRedditPoller_Poll_InvalidXML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid xml`))
	}))
	defer srv.Close()

	p := &RedditPoller{client: srv.Client()}
	_, err := p.Poll(context.Background(), srv.URL+"/r/golang", time.Now())
	require.Error(t, err)

	var pollerErr *Error
	require.ErrorAs(t, err, &pollerErr)
	assert.ErrorIs(t, pollerErr, ErrParse)
}

func TestYouTubePoller_Source(t *testing.T) {
	p := NewYouTubePoller("key")
	assert.Equal(t, "youtube", p.Source())
}

func TestYouTubePoller_Supports(t *testing.T) {
	p := NewYouTubePoller("key")
	assert.True(t, p.Supports("https://youtube.com/@GolangCafe"))
	assert.True(t, p.Supports("https://www.youtube.com/@GolangCafe"))
	assert.False(t, p.Supports("https://github.com/golang/go"))
}

func TestYouTubePoller_Poll_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": [{"snippet": {
			"title": "Go 1.23 Features",
			"publishedAt": "2099-01-01T00:00:00Z",
			"resourceId": {"videoId": "abc123"}
		}}]}`))
	}))
	defer srv.Close()

	p := NewYouTubePoller("key")
	p.client = srv.Client()
	p.apiURL = srv.URL

	events, err := p.Poll(context.Background(), "https://youtube.com/@GolangCafe", time.Now().Add(-time.Hour))
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "youtube", events[0].Source)
	assert.Equal(t, "Go 1.23 Features", events[0].Title)
	assert.Contains(t, events[0].Link, "abc123")
}

func TestYouTubePoller_Poll_FiltersOldVideos(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": [{"snippet": {
			"title": "Old video",
			"publishedAt": "2000-01-01T00:00:00Z",
			"resourceId": {"videoId": "abc123"}
		}}]}`))
	}))
	defer srv.Close()

	p := NewYouTubePoller("key")
	p.client = srv.Client()
	p.apiURL = srv.URL

	events, err := p.Poll(context.Background(), "https://youtube.com/@GolangCafe", time.Now())
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestYouTubePoller_Poll_InvalidURL(t *testing.T) {
	p := NewYouTubePoller("key")
	_, err := p.Poll(context.Background(), "https://youtube.com/", time.Now())
	require.Error(t, err)

	var pollerErr *Error
	require.ErrorAs(t, err, &pollerErr)
	assert.ErrorIs(t, pollerErr, ErrInvalidURL)
}

func TestYouTubePoller_Poll_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	p := NewYouTubePoller("key")
	p.client = srv.Client()
	p.apiURL = srv.URL

	_, err := p.Poll(context.Background(), "https://youtube.com/@GolangCafe", time.Now())
	require.Error(t, err)

	var pollerErr *Error
	require.ErrorAs(t, err, &pollerErr)
	assert.ErrorIs(t, pollerErr, ErrFetch)
}

func TestYouTubePoller_Poll_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer srv.Close()

	p := NewYouTubePoller("key")
	p.client = srv.Client()
	p.apiURL = srv.URL

	_, err := p.Poll(context.Background(), "https://youtube.com/@GolangCafe", time.Now())
	require.Error(t, err)

	var pollerErr *Error
	require.ErrorAs(t, err, &pollerErr)
	assert.ErrorIs(t, pollerErr, ErrParse)
}

func TestBuildRedditFeedURL(t *testing.T) {
	assert.Equal(t, "https://reddit.com/r/golang/.rss", buildRedditFeedURL("https://reddit.com/r/golang"))
	assert.Equal(t, "https://reddit.com/r/golang/.rss", buildRedditFeedURL("https://reddit.com/r/golang/"))
	assert.Equal(t, "https://reddit.com/r/golang/.rss", buildRedditFeedURL("https://reddit.com/r/golang/.rss"))
}

func TestExtractSOTag(t *testing.T) {
	tag, err := extractSOTag("https://stackoverflow.com/questions/tagged/golang")
	require.NoError(t, err)
	assert.Equal(t, "golang", tag)

	_, err = extractSOTag("https://stackoverflow.com/questions/tagged/")
	require.Error(t, err)
}

func TestExtractYouTubeChannelID(t *testing.T) {
	id, err := extractYouTubeChannelID("https://youtube.com/@GolangCafe")
	require.NoError(t, err)
	assert.Equal(t, "GolangCafe", id)

	id, err = extractYouTubeChannelID("https://www.youtube.com/@GolangCafe")
	require.NoError(t, err)
	assert.Equal(t, "GolangCafe", id)

	_, err = extractYouTubeChannelID("https://youtube.com/")
	require.Error(t, err)
}

func TestScheduler_Watch_Unwatch(t *testing.T) {
	s := NewScheduler(nil, time.Minute, newTestLogger())

	s.Watch("https://reddit.com/r/golang")
	s.Watch("https://reddit.com/r/golang")

	s.mu.RLock()
	assert.Len(t, s.watching, 1)
	s.mu.RUnlock()

	s.Unwatch("https://reddit.com/r/golang")

	s.mu.RLock()
	assert.Empty(t, s.watching)
	s.mu.RUnlock()
}

func TestScheduler_FindPoller(t *testing.T) {
	so := NewStackOverflowPoller()
	reddit := NewRedditPoller()
	youtube := NewYouTubePoller("key")

	s := NewScheduler(nil, time.Minute, newTestLogger(), so, reddit, youtube)

	assert.Equal(t, so, s.findPoller("https://stackoverflow.com/questions/tagged/golang"))
	assert.Equal(t, reddit, s.findPoller("https://reddit.com/r/golang"))
	assert.Equal(t, youtube, s.findPoller("https://youtube.com/@GolangCafe"))
	assert.Nil(t, s.findPoller("https://github.com/golang/go"))
}

func TestSourceToTopic(t *testing.T) {
	assert.Equal(t, "events.answer", sourceToTopic("stackoverflow").String())
	assert.Equal(t, "events.post", sourceToTopic("reddit").String())
	assert.Equal(t, "events.video", sourceToTopic("youtube").String())
	assert.Equal(t, "events.push", sourceToTopic("unknown").String())
}

func TestSourceToEventType(t *testing.T) {
	assert.Equal(t, "answer", sourceToEventType("stackoverflow"))
	assert.Equal(t, "post", sourceToEventType("reddit"))
	assert.Equal(t, "video", sourceToEventType("youtube"))
	assert.Equal(t, "push", sourceToEventType("unknown"))
}
