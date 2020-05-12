BINPATH ?= build
BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

all: audit test build

audit:
	nancy go.sum

build:
	go build -tags 'production' -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)" -o $(BINPATH)/dp-frontend-geography-controller

debug:
	go build -tags 'debug' -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)" -o $(BINPATH)/dp-frontend-geography-controller
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontend-geography-controller

test:
	go test -race -cover -tags 'production' ./...

convey:
	goconvey ./...

.PHONY: all audit build debug
