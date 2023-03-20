NAME = zerok-injector
IMAGE_PREFIX = us-west1-docker.pkg.dev/zerok-dev/stage
IMAGE_NAME = zerok-injector
IMAGE_VERSION = 0.0.2

export GO111MODULE=on

build: sync
	go build -v -o $(NAME) cmd/main.go

sync:
	go get -v ./...
	
docker-build:
	docker build --no-cache -t $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION) .
	
docker-push:
	docker push $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION) 

kind:
	kind create cluster --config kind.yaml