.PHONY: build test test-integration generate install-paramgen install-tools golangci-lint-install

VERSION=$(shell git describe --tags --dirty --always)

build:
	go build -ldflags "-X 'github.com/alarbada/conduit-connector-rabbitmq.version=${VERSION}'" -o conduit-connector-rabbitmq cmd/connector/main.go

test:
	docker compose -f test/docker-compose.yml up --quiet-pull -d --wait 
	go test -count=1 -v -race .; ret=$$?; \
		docker compose -f test/docker-compose.yml down; \
		exit $$ret

generate:
	go generate ./...

install-paramgen:
	go install github.com/conduitio/conduit-connector-sdk/cmd/paramgen@latest

lint:
	golangci-lint run

acceptance:
	go test -v -race -count=1 -run TestAcceptance .

up:
	docker compose -f test/docker-compose.yml up --quiet-pull -d --wait 

down:
	docker compose -f test/docker-compose.yml down -v --remove-orphans

