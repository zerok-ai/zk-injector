NAME = zk-injector
IMAGE_PREFIX = rajeevzerok
IMAGE_NAME = zk-injector
IMAGE_VERSION = 0.5

export GO111MODULE=on

app: deps
	go build -v -o $(NAME) cmd/main.go

deps:
	go get -v ./...
	
docker:
	docker build --no-cache -t $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION) .
	
push:
	docker push $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION) 

kind:
	kind create cluster --config kind.yaml