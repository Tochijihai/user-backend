APP_NAME = go-lambda-app
BUILDER_NAME = go-lambda-builder
CONTAINER_NAME = go-lambda-builder-container

.PHONY: build clean zip extract build-UserBackendFunction

build:
	docker build -t $(APP_NAME) .

# SAM用のビルドターゲット - 既にビルドされたbootstrapファイルをコピー
build-UserBackendFunction:
	cp bootstrap bootstrap || echo "bootstrap file already exists"

build-builder:
	docker build -t $(BUILDER_NAME) --target builder .

clean:
	docker rmi -f $(APP_NAME) || true
	docker rmi -f $(BUILDER_NAME) || true

zip: build-builder
	docker run --rm $(BUILDER_NAME) sh -c "apk add --no-cache zip && zip function.zip bootstrap"

extract:
	docker create --name $(CONTAINER_NAME) $(BUILDER_NAME)
	docker cp $(CONTAINER_NAME):/app/function.zip .
	docker rm $(CONTAINER_NAME)
