BINPATH ?= build

build:
	go build -tags 'production' -o $(BINPATH)/dp-frontend-geography-controller

debug:
	go build -tags 'debug' -o $(BINPATH)/dp-frontend-geography-controller
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontend-geography-controller

test:
	go test -cover $(shell go list ./... | grep -v /vendor/) -tags 'production' ./...

convey:
	goconvey ./...

.PHONY: build debug
