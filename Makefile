PROJECTNAME := $(shell basename "$(PWD)")

## build: clean and build binary file
.PHONY: build
build:
	go build ./...

## install: get packages
.PHONY: install
install:
	go mod download

## test: go run test
.PHONY: test
test:
	go test -race ./...

## lint: format code as golangci-lint
.PHONY: lint
lint:
	golangci-lint run ./...

## image: make docker image
.PHONY: image
image:
	docker build -t orchid .

## up: docker compose and start up
.PHONY: up
up: down
	docker-compose pull orchid
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 docker-compose up --build -d

## down: stop
.PHONY: down
down:
	docker-compose down --remove-orphans

## logs: print out logs
.PHONY: logs
logs:
	docker-compose logs -f orchid
	
.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
