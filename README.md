# GitNotify

Real-time notifications from GitHub, GitLab, Stack Overflow, Reddit and YouTube — delivered to Telegram.

## How it works
```
GitHub/GitLab ──webhook──► Webhook Receiver ──►
                                                  Kafka ──► Notifier ──► Telegram
SO/Reddit/YouTube ─poll──► Poller ───────────►
```

Users subscribe to repositories and resources via Telegram bot. When something happens — a push, pull request, new answer or video — they get a message.

## Services

| Service | Description |
|---|---|
| `bot` | Telegram bot — manages subscriptions |
| `webhook` | HTTP server — receives GitHub/GitLab webhooks |
| `poller` | Scheduler — polls Stack Overflow, Reddit, YouTube |
| `notifier` | Kafka consumer — sends Telegram notifications |

## Supported sources

| Source | How | Events |
|---|---|---|
| GitHub | Webhook | push, pull request, issue, pipeline |
| GitLab | Webhook | push, merge request, issue, pipeline |
| Stack Overflow | Polling | new questions by tag |
| Reddit | RSS polling | new posts in subreddit |
| YouTube | API polling | new videos on channel |

## Stack

- **Go 1.23**
- **Apache Kafka** — event bus between services
- **PostgreSQL** — bot stores subscriptions, notifier keeps its own copy
- **Telegram Bot API**
- **Docker / Docker Compose**

## Quick start

### 1. Clone and configure
```bash
git clone https://github.com/dlc-01/GitNotify
cd GitNotify
cp .env.bot.example .env.bot
cp .env.webhook.example .env.webhook
cp .env.notifier.example .env.notifier
cp .env.poller.example .env.poller
```

Fill in the tokens:

**.env.bot**
```
BOT_TOKEN=your_token_from_botfather
POSTGRES_HOST=postgres-bot
POSTGRES_PORT=5432
POSTGRES_USER=gitnotify
POSTGRES_PASSWORD=gitnotify
POSTGRES_DBNAME=gitnotify_bot
KAFKA_BROKERS=kafka:9092
```

**.env.webhook**
```
WEBHOOK_HOST=0.0.0.0
WEBHOOK_PORT=8080
WEBHOOK_GITHUBSECRET=your_github_webhook_secret
WEBHOOK_GITLABSECRET=your_gitlab_webhook_secret
KAFKA_BROKERS=kafka:9092
```

**.env.notifier**
```
BOT_TOKEN=your_token_from_botfather
POSTGRES_HOST=postgres-notifier
POSTGRES_PORT=5432
POSTGRES_USER=gitnotify
POSTGRES_PASSWORD=gitnotify
POSTGRES_DBNAME=gitnotify_notifier
KAFKA_BROKERS=kafka:9092
```

**.env.poller**
```
KAFKA_BROKERS=kafka:9092
POLLER_YOUTUBEAPIKEY=your_youtube_api_key
```

### 2. Run
```bash
docker compose up -d
```

### 3. Set up webhooks

Go to your GitHub repository → Settings → Webhooks → Add webhook:
- **Payload URL**: `https://your-domain.com/webhook`
- **Content type**: `application/json`
- **Secret**: value from `WEBHOOK_GITHUBSECRET`
- **Events**: push, pull requests, issues

Same for GitLab: Settings → Webhooks.

## Bot commands
```
/subscribe <url>       subscribe to a repository or resource
/unsubscribe <url>     unsubscribe
/list                  list your subscriptions
/mute <url> <event>    mute an event type
```

**Event types:** `push` `pr` `issue` `pipeline` `answer` `post` `video`

**Examples:**
```
/subscribe https://github.com/golang/go
/subscribe https://stackoverflow.com/questions/tagged/golang
/subscribe https://reddit.com/r/golang
/subscribe https://youtube.com/@GolangCafe
/mute https://github.com/golang/go push
```

## Project structure
```
GitNotify/
├── cmd/
│   ├── bot/           entry point — Telegram bot
│   ├── webhook/       entry point — webhook receiver
│   ├── notifier/      entry point — Kafka consumer
│   └── poller/        entry point — polling scheduler
├── internal/
│   ├── bot/           bot handlers, commands, callbacks
│   ├── config/        configuration (yaml + env)
│   ├── domain/        shared types
│   ├── kafka/         producer, consumer, topics, messages
│   ├── notifier/      notification logic, repository
│   ├── poller/        polling logic for external sources
│   ├── repository/    bot repository (postgres)
│   └── webhook/       webhook handler, parser, validator
└── migrations/
    ├── bot/
    └── notifier/
```

## Running tests
```bash
# unit tests
go test ./...

# integration tests (requires Docker)
go test -tags integration ./internal/repository/postgres/... -v
go test -tags integration ./internal/notifier/... -v
```

## Architecture decisions

**Why Kafka?** Each service has its own database and communicates only through Kafka topics. The bot produces subscription events, the notifier consumes them and keeps its own copy. This way services are fully decoupled — you can add a Discord notifier without touching the bot.

**Why separate databases?** The bot owns user and chat data. The notifier only needs a flat list of `(chat_id, repo_url, muted_events)` — no users, no chats. Separate schemas make each service independently deployable.

**Why polling for SO/Reddit/YouTube?** These platforms don't support webhooks for arbitrary content. Stack Overflow and Reddit are free with no auth required. YouTube needs an API key but offers 10k free requests per day — enough for reasonable polling intervals.
## Try it

Telegram bot is live: [@NotificationCollect_bot](https://t.me/NotificationCollect_bot)

## Exposing webhook locally

The webhook receiver needs a public URL. Use [ngrok](https://ngrok.com) for local development:
```bash
ngrok http 8080
```

You'll get a URL like `https://abc123.ngrok.io` — use it as the Payload URL in GitHub/GitLab webhook settings.

Note: ngrok URL changes on every restart. For permanent hosting use a VPS or a platform like Railway or Render.

## GitHub webhook setup

1. Go to your repository → **Settings** → **Webhooks** → **Add webhook**
2. Set **Payload URL** to `https://your-domain.com/webhook`
3. Set **Content type** to `application/json`
4. Set **Secret** — use the same value as `WEBHOOK_GITHUBSECRET` in `.env.webhook`
5. Choose events: `push`, `pull requests`, `issues`
6. Click **Add webhook**

## GitLab webhook setup

1. Go to your project → **Settings** → **Webhooks**
2. Set **URL** to `https://your-domain.com/webhook`
3. Set **Secret token** — use the same value as `WEBHOOK_GITLABSECRET` in `.env.webhook`
4. Choose triggers: `Push events`, `Merge request events`, `Issues events`
5. Click **Add webhook**