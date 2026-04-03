.PHONY: gen-icon build install run

gen-icon:
	go run ./cmd/gen-icon

build:
	go build -o gl1tch-notify .

install:
	go install .
	ln -sf $(shell go env GOPATH)/bin/gl1tch-notify $(HOME)/.local/bin/gl1tch-notify

run:
	go run .
