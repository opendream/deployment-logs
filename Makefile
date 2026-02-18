IMAGE := rutcreate/deployment-logs
TAG := latest

.PHONY: setup build push

setup:
	docker buildx create --name multiplatform --use || true

build:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMAGE):$(TAG) --push .

push: build
