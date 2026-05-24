module github.com/Sushka21/microservices-ecommerce/notifications

go 1.26.2

require (
	github.com/caarlos0/env/v10 v10.0.0
	github.com/Sushka21/microservices-ecommerce/pkg v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.9.2
	github.com/pressly/goose/v3 v3.26.0
	github.com/stretchr/testify v1.11.1
	go.uber.org/mock v0.6.0
	go.uber.org/zap v1.28.0
	golang.org/x/sync v0.20.0
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	golang.org/x/oauth2 v0.35.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260217215200-42d3e9bedb6d // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.5.3
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace github.com/Sushka21/microservices-ecommerce/pkg => ../pkg
