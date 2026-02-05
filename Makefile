.PHONY: build

build: 
		go build -v ./cmd/apiserver

start:
		./apiserver

.DEFAULT_GOAL := build