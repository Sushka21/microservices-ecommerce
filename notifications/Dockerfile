FROM golang:1.25 AS builder

WORKDIR /app

COPY go.work ./
COPY notifications/go.mod notifications/go.sum ./notifications/
COPY pkg/go.mod pkg/go.sum ./pkg/

COPY . .

WORKDIR /app/notifications
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/notifications ./cmd/app

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/notifications /app/notifications

CMD ["./notifications"]
