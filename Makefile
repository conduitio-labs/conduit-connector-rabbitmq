VERSION=$(shell git describe --tags --dirty --always)

.PHONY: build
build:
	go build -ldflags "-X 'github.com/conduitio-labs/conduit-connector-rabbitmq.version=${VERSION}'" -o conduit-connector-rabbitmq cmd/connector/main.go

.PHONY: test
test:
	docker compose -f test/docker-compose.yml up --quiet-pull -d --wait 
	go test -count=1 -v -race .; ret=$$?; \
		docker compose -f test/docker-compose.yml down; \
		exit $$ret

.PHONY: generate
generate:
	go generate ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: acceptance
acceptance:
	go test -v -race -count=1 -run TestAcceptance .

.PHONY: up
up:
	docker compose -f test/docker-compose.yml up --quiet-pull -d --wait 

.PHONY: down
down:
	docker compose -f test/docker-compose.yml down -v --remove-orphans

.PHONY: install-tools
install-tools:
	@echo Installing tools from tools.go
	@go list -e -f '{{ join .Imports "\n" }}' tools.go | xargs -I % go list -f "%@{{.Module.Version}}" % | xargs -tI % go install %
	@go mod tidy
