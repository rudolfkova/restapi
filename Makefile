.PHONY: build

build: 
		go build -v ./cmd/apiserver

.PHONY: start

start:
		./apiserver

.PHONY: test

test:
	go test -v -race -timeout 30s ./...

.PHONY: db

db:
	psql -U postgres -d restapi_dev

.DEFAULT_GOAL := build