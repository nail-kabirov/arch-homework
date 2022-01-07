all: modules build

AUTH_DOCKER_IMAGE_NAME=nailkabirov/arch-hw7-auth
AUTH_DOCKER_IMAGE_TAG=v0.0.1

USER_DOCKER_IMAGE_NAME=nailkabirov/arch-hw7-user
USER_DOCKER_IMAGE_TAG=v0.0.1

ORDER_DOCKER_IMAGE_NAME=nailkabirov/arch-hw7-order
ORDER_DOCKER_IMAGE_TAG=v0.0.1

BILLING_DOCKER_IMAGE_NAME=nailkabirov/arch-hw7-billing
BILLING_DOCKER_IMAGE_TAG=v0.0.1

NOTIFICATION_DOCKER_IMAGE_NAME=nailkabirov/arch-hw7-notification
NOTIFICATION_DOCKER_IMAGE_TAG=v0.0.1

APP_NAMES = \
	auth \
	billing \
	notification \
	order \
	user

.PHONY: modules
modules:
	go mod tidy

.PHONY: build
build: modules check build_auth build_billing build_notification build_order build_user

.PHONY: build_auth
build_auth:
	docker build -t $(AUTH_DOCKER_IMAGE_NAME):$(AUTH_DOCKER_IMAGE_TAG) -f Auth.Dockerfile .

.PHONY: build_user
build_user:
	docker build -t $(USER_DOCKER_IMAGE_NAME):$(USER_DOCKER_IMAGE_TAG) -f User.Dockerfile .

.PHONY: build_order
build_order:
	docker build -t $(ORDER_DOCKER_IMAGE_NAME):$(ORDER_DOCKER_IMAGE_TAG) -f Order.Dockerfile .

.PHONY: build_notification
build_notification:
	docker build -t $(NOTIFICATION_DOCKER_IMAGE_NAME):$(NOTIFICATION_DOCKER_IMAGE_TAG) -f Notification.Dockerfile .

.PHONY: build_billing
build_billing:
	docker build -t $(BILLING_DOCKER_IMAGE_NAME):$(BILLING_DOCKER_IMAGE_TAG) -f Billing.Dockerfile .

.PHONY: build_local
build_local: modules check
	$(foreach name, $(APP_NAMES), CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bin/$(name) ./cmd/$(name)/;)

.PHONY: clean
clean:
	rm -f $(foreach name,$(APP_NAMES), "bin/$(name)")

.PHONY: check
check:
	golangci-lint run