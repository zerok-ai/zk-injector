NAME = zk-injector
IMAGE_PREFIX = rajeevzerok
IMAGE_NAME = zk-injector
IMAGE_VERSION = 0.6

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