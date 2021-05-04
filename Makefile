PROJECT_NAME := api-server-poc
API_IMAGE_NAME := $(PROJECT_NAME)-api
API_CONTAINER_NAME := $(API_IMAGE_NAME)-container
NETWORK_NAME := $(PROJECT_NAME)-network

.PHONY: fmt
fmt:
	go fmt ./...
	make -C spec fmt

.PHONY: build
build:
	docker build -t $(API_IMAGE_NAME):latest .

.PHONY: start
start:
	if [ -z "`docker network ls | grep $(NETWORK_NAME)`" ]; then \
		docker network create $(NETWORK_NAME); \
	fi
	if [ -z "`docker ps | grep $(API_CONTAINER_NAME)`" ]; then \
		docker run -d --rm --name $(API_CONTAINER_NAME) --network $(NETWORK_NAME) -p 8080:8080 $(API_IMAGE_NAME):latest; \
	fi


.PHONY: stop
stop:
	if [ -n "`docker ps | grep $(API_CONTAINER_NAME)`" ]; then \
		docker stop $(API_CONTAINER_NAME); \
	fi
	if [ -n "`docker network ls | grep $(NETWORK_NAME)`" ]; then \
		docker network rm $(NETWORK_NAME); \
	fi

.PHONY: spec
spec: start
	make -C spec run
