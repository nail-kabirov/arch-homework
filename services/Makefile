all: modules build

AUTH_DOCKER_IMAGE_NAME=nailkabirov/arch-hw5-auth
AUTH_DOCKER_IMAGE_TAG=v0.0.2

USER_DOCKER_IMAGE_NAME=nailkabirov/arch-hw5-user
USER_DOCKER_IMAGE_TAG=v0.0.2

APP_NAMES = \
	auth \
	user

.PHONY: modules
modules:
	go mod tidy

.PHONY: build
build: modules check build_auth build_user

.PHONY: build_auth
build_auth:
	docker build -t $(AUTH_DOCKER_IMAGE_NAME):$(AUTH_DOCKER_IMAGE_TAG) -f Auth.Dockerfile .

.PHONY: build_user
build_user:
	docker build -t $(USER_DOCKER_IMAGE_NAME):$(USER_DOCKER_IMAGE_TAG) -f User.Dockerfile .

.PHONY: build_local
build_local: modules check
	$(foreach name, $(APP_NAMES), CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bin/$(name) ./cmd/$(name)/;)

.PHONY: clean
clean:
	rm -f $(foreach name,$(APP_NAMES), "bin/$(name)")

.PHONY: check
check:
	golangci-lint run