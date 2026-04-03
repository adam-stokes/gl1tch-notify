.PHONY: gen-icon build install run

gen-icon:
	go run ./cmd/gen-icon

build:
	go build -o gl1tch-notify .

install:
	go install .

run:
	go run .
