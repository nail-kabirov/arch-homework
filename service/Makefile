all: modules build

DOCKER_IMAGE_NAME=nailkabirov/arch-hw3
DOCKER_IMAGE_TAG=v0.0.4

.PHONY: modules
modules:
	go mod tidy

.PHONY: build
build:
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

.PHONY: build_local
build_local: modules check
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bin/user ./cmd/user/

.PHONY: clean
clean:
	rm -f "./bin/user")

.PHONY: check
check:
	golangci-lint run