# GitNotify

Real-time notifications from GitHub, GitLab, Stack Overflow, Reddit and YouTube — delivered to Telegram.

Try it: [@NotificationCollect_bot](https://t.me/NotificationCollect_bot)

## How it works
```
User → /subscribe → Bot → Kafka (subscriptions.created)
                               │
                               ▼
                            Poller
                  polls GitHub · GitLab · SO · Reddit · YouTube
                               │
                               ▼
                      Kafka (events.push/pr/issue/...)
                               │
                               ▼
                            Notifier
                               │
                               ▼
                            Telegram
```

Users subscribe to any public repository or resource via Telegram bot. The poller periodically checks for new events and sends them to Kafka. The notifier picks them up and delivers to Telegram.

## Services

| Service | Description |
|---|---|
| `bot` | Telegram bot — manages subscriptions |
| `poller` | Scheduler — polls all sources via API |
| `notifier` | Kafka consumer — sends Telegram notifications |

## Supported sources

| Source | How | Events |
|---|---|---|
| GitHub | REST API polling | push, pull request, issue |
| GitLab | REST API polling | push, merge request, issue |
| Stack Overflow | API polling | new questions by tag |
| Reddit | RSS polling | new posts in subreddit |
| YouTube | Data API v3 | new videos on channel |

## Stack

- **Go 1.23**
- **Apache Kafka** — event bus between services
- **PostgreSQL** — bot stores subscriptions, notifier keeps its own copy
- **Telegram Bot API**
- **Docker / Docker Compose**

## Quick start

### 1. Clone
```bash
git clone https://github.com/dlc-01/GitNotify
cd GitNotify
```

### 2. Create env files

**.env.bot**
```env
BOT_TOKEN=your_token_from_botfather
POSTGRES_HOST=postgres-bot
POSTGRES_PORT=5432
POSTGRES_USER=gitnotify
POSTGRES_PASSWORD=gitnotify
POSTGRES_DBNAME=gitnotify_bot
KAFKA_BROKERS=kafka:9092
DEBUG=false
```

**.env.notifier**
```env
BOT_TOKEN=your_token_from_botfather
POSTGRES_HOST=postgres-notifier
POSTGRES_PORT=5432
POSTGRES_USER=gitnotify
POSTGRES_PASSWORD=gitnotify
POSTGRES_DBNAME=gitnotify_notifier
KAFKA_BROKERS=kafka:9092
DEBUG=false
```

**.env.poller**
```env
KAFKA_BROKERS=kafka:9092
POLLER_INTERVAL=5m
POLLER_GITHUBTOKEN=your_github_personal_access_token
POLLER_GITLABTOKEN=your_gitlab_personal_access_token
POLLER_YOUTUBEAPIKEY=your_youtube_data_api_key
DEBUG=false
```

### 3. Get API tokens

**GitHub token** — increases rate limit from 60 to 5000 req/hour:
1. GitHub → Settings → Developer settings → Personal access tokens → Generate new token
2. Permissions: `Contents` read-only, `Metadata` read-only
3. No scopes needed for public repositories

**GitLab token** — optional, needed only for private repos:
1. GitLab → User Settings → Access Tokens → Create token
2. Scope: `read_api`

**YouTube API key** — required:
1. [console.cloud.google.com](https://console.cloud.google.com) → Create project
2. APIs & Services → Enable → YouTube Data API v3
3. Credentials → Create Credentials → API key

Stack Overflow and Reddit work without any tokens.

### 4. Run
```bash
docker compose up -d
```

### 5. Try it

Open [@NotificationCollect_bot](https://t.me/NotificationCollect_bot) and try:
```
/sources
/subscribe https://github.com/golang/go
/subscribe https://stackoverflow.com/questions/tagged/golang
/subscribe https://reddit.com/r/golang
/subscribe https://youtube.com/@GolangCafe
/list
/mute https://github.com/golang/go push
/unmute https://github.com/golang/go push
```

## Bot commands

| Command | Description |
|---|---|
| `/subscribe <url>` | Subscribe to a repository or resource |
| `/unsubscribe <url>` | Unsubscribe |
| `/list` | List your subscriptions with inline controls |
| `/mute <url> <event>` | Mute an event type |
| `/unmute <url> <event>` | Unmute an event type |
| `/sources` | Show supported sources and event types |
| `/help` | Show all commands |

**Event types:** `push` `pr` `issue` `pipeline` `answer` `post` `video`

## Project structure
```
GitNotify/
├── cmd/
│   ├── bot/             entry point — Telegram bot
│   ├── notifier/        entry point — Kafka consumer
│   └── poller/          entry point — polling scheduler
├── internal/
│   ├── bot/             bot handlers, commands, callbacks
│   │   ├── core/        registry, sender, middleware
│   │   ├── commands/    subscribe, unsubscribe, list, mute, unmute, sources, help
│   │   └── callback/    inline keyboard handlers
│   ├── config/          configuration (env + optional yaml)
│   ├── domain/          shared types — User, Chat, Subscription, EventType
│   ├── kafka/           producer, multi-producer, consumer, topics, messages
│   ├── notifier/        notification handler, postgres repository
│   ├── poller/          GitHub, GitLab, SO, Reddit, YouTube pollers + scheduler
│   └── repository/      bot postgres repository + logging decorator
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

**Why Kafka?** Each service has its own database and communicates only through Kafka topics. The bot produces subscription events, the notifier consumes them and keeps its own copy. Adding a Discord notifier or Slack integration requires zero changes to existing services.

**Why separate databases?** The bot owns user and chat data. The notifier only needs a flat list of `(chat_id, repo_url, muted_events)`. Separate schemas make each service independently deployable and scalable.

**Why polling instead of webhooks?** Webhooks require a public URL and manual setup per repository. Polling via API works for any public repository without configuration — users just paste a URL and subscribe. GitHub API allows 5000 requests per hour with a token, which is sufficient for reasonable polling intervals.

**Why groups support?** Telegram groups are a natural fit for team notifications. Any admin can subscribe a group chat to a repository, and the whole team gets notified without each member setting up individual subscriptions.