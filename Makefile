NAME = asm-webhook
IMAGE_NAME = asm-webhook
IMAGE_PREFIX = ayoul3
BUILD=go build -ldflags="-s -w"
IMAGE_VERSION = $$(git log --abbrev-commit --format=%h -s | head -n 1)
export GO111MODULE=on

test:
	go test -v ./... -cover

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(BUILD) -o asm-webhook main.go

docker:
	docker build --no-cache -t $(IMAGE_PREFIX)/$(IMAGE_NAME):latest .

push: docker
	docker push $(IMAGE_PREFIX)/$(IMAGE_NAME):latest

