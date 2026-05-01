FROM golang:1.25

WORKDIR /tests

COPY integration-tests/go.mod integration-tests/go.sum ./
RUN go mod download

COPY integration-tests/. .
RUN go mod tidy

CMD ["go", "test", "-v", "./..."]
