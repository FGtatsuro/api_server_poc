PROJECT_NAME := api-server-poc
API_IMAGE_NAME := $(PROJECT_NAME)-api
API_CONTAINER_NAME := $(API_IMAGE_NAME)-container
SPEC_IMAGE_NAME := $(PROJECT_NAME)-spec
NETWORK_NAME := $(PROJECT_NAME)-network

.PHONY: fmt
fmt:
	pip3 install -r requirements-dev.txt
	autopep8 --in-place --aggressive --recursive .
	isort .

.PHONY: build
build:
	docker build -t $(SPEC_IMAGE_NAME):latest .

.PHONY: run
run: build
	docker run -t --rm --network $(NETWORK_NAME) --env API_URL="http://$(API_CONTAINER_NAME):8080" $(SPEC_IMAGE_NAME):latest
