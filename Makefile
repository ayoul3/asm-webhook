NAME = ssm-webhook
IMAGE_NAME = ssm-webhook
IMAGE_PREFIX = ayoul3
IMAGE_VERSION = $$(git log --abbrev-commit --format=%h -s | head -n 1)
export GO111MODULE=on

test:
	go test -v ./... -cover

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ssm-webhook main.go

docker: build
	#docker build --no-cache -t $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION) .
	docker build --no-cache -t $(IMAGE_PREFIX)/$(IMAGE_NAME):latest .
	#docker tag $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION) $(IMAGE_PREFIX)/$(IMAGE_NAME):latest

push: docker
	#docker push $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION)
	docker push $(IMAGE_PREFIX)/$(IMAGE_NAME):latest

