.PHONY: all setup test build install

all: build test install

setup:
	go get -u -v

test:
	go test

build:
	go build -i

install:
	go install
