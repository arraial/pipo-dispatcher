APP=pipo_dispatcher
PRINT=@echo

-include .env

define help_message
Usage:
help          show this message
test_setup    install test dependencies
dev_setup     install dev dependencies
lint          run dev utilities for code quality assurance
format        run dev utilities for code format assurance
docs          generate code documentation
test          run test suite
dist          package application for distribution
image         build app docker image
test_image    run test suite in a container
run_image     run app docker image in a container
endef
export help_message

.PHONY: help
help:
	$(PRINT) "$$help_message"

.PHONY: dev_setup
dev_setup:
	go mod download

.PHONY: test_setup
test_setup: dev_setup

.PHONY: update_deps
update_deps:
	go get -u
	go mod tidy

.PHONY: lint
lint:
	go vet ./...

.PHONY: format
format:
	go fmt ./...

.PHONY: test
test:
	go test ./...

.PHONY: docs
docs:
	go doc --all ./...

.PHONY: dist
dist:
	go build -o ./build/app ./cmd

.PHONY: image
image:
	docker buildx bake image

.PHONY: test_image
test_image:
	docker buildx bake test

.PHONY: run_image
run_image: image
	docker run -d --name $(APP) --env-file .env $(APP):latest
