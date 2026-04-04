.PHONY: gen-icon build install run

gen-icon:
	go run ./cmd/gen-icon

build:
	go build -o glitch-notify .

install:
	go install .
	ln -sf $(shell go env GOPATH)/bin/glitch-notify $(HOME)/.local/bin/glitch-notify

run:
	go run .
