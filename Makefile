APP_NAME := blog-api
DOCKER_IMAGE := blog-api-image
DOCKER_TAG := latest

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo
	@echo "Targets:"
	@awk '/^[a-zA-Z\-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: run
run:
	go run main.go

.PHONY: build
build:
	go build -o bin/$(APP_NAME) main.go

.PHONY: test
test:
	go test ./... -v

.PHONY: unit-test
unit-test:
	go test ./internal/... -v

.PHONY: integration-test
integration-test:
	go test ./tests/integration_test.go -v

.PHONY: fuzz-test
fuzz-test:
	go test ./tests/fuzz_test.go -v

.PHONY: lint
lint:
	golangci-lint run

.PHONY: format
format:
	gofmt -w .

.PHONY: clean
clean:
	rm -rf bin

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run:
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-clean
docker-clean:
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG)
