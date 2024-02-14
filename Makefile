.PHONY: build test test-integration generate install-paramgen install-tools golangci-lint-install

VERSION=$(shell git describe --tags --dirty --always)

build:
	go build -ldflags "-X 'github.com/alarbada/conduit-connector-rabbitmq.version=${VERSION}'" -o rabbitmq cmd/connector/main.go

test:
	# run required docker containers, execute integration tests, stop containers after tests
	docker compose -f test/docker-compose.yml up --quiet-pull -d --wait 
	go test $(GOTEST_FLAGS) -v -race ./...; ret=$$?; \
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

up-tls:
	docker compose -f test/docker-compose-tls.yml up --quiet-pull -d --wait 


down-tls:
	docker compose -f test/docker-compose-tls.yml down -v --remove-orphans

clean:
	rm -rf test/*.pem
