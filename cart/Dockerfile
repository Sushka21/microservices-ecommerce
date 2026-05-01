FROM golang:1.25 AS builder

WORKDIR /app

COPY go.work ./
COPY cart/go.mod cart/go.sum ./cart/
COPY pkg/go.mod pkg/go.sum ./pkg/

COPY . .

WORKDIR /app/cart
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/cart ./cmd/app

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/cart /app/cart

CMD ["./cart"]
