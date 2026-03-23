FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o notifier ./cmd/notifier

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/notifier .
COPY --from=builder /app/migrations ./migrations
CMD ["./notifier"]