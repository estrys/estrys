.PHONY: run
run:
	docker-compose up

.PHONY: install-linter
install-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1

.PHONY: install-tools
install-tools: install-linter
	go install github.com/cosmtrek/air@v1.40.4
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2
	go install github.com/volatiletech/sqlboiler/v4@latest
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
	go install github.com/daixiang0/gci@v0.9.0
	go install github.com/vektra/mockery/v2@v2.15.0


.PHONY: db-migrate
db-migrate:
	docker-compose run --entrypoint /bin/bash estrys -c 'set -o allexport; source .env; set +o allexport; migrate -path ./migrations/ -database $$DB_URL up'

.PHONY: db-drop
db-drop:
	docker-compose run --entrypoint /bin/bash estrys -c 'set -o allexport; source .env; set +o allexport; migrate -path ./migrations/ -database $$DB_URL drop'

.PHONY: models
models:
	docker-compose run --entrypoint /bin/bash estrys -c 'set -o allexport; source .env; set +o allexport; sqlboiler psql'


.PHONY: format
format:
	gci write --skip-generated -s "standard,prefix(github.com/estrys/estrys),default" ./

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -v -race -coverprofile=coverage.out -coverpkg=./internal/... ./internal/...

.PHONY: build
build:
	go build -o estrys ./cmd/estrys/
	go build -o worker ./cmd/worker/

.PHONY: generate
generate:
	go generate ./...
