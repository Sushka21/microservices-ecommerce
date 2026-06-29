FROM golang:1.26.2 AS builder

WORKDIR /app

COPY go.work ./
COPY notifications/go.mod notifications/go.sum ./notifications/
COPY pkg/go.mod pkg/go.sum ./pkg/

COPY . .

WORKDIR /app/notifications
RUN go mod download

# Go client is based on the C library librdkafka
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/notifications ./cmd/app

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /bin/notifications /app/notifications

CMD ["./notifications"]
