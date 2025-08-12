APP_NAME = go-lambda-app
BUILDER_NAME = go-lambda-builder
CONTAINER_NAME = go-lambda-builder-container

.PHONY: build clean zip extract build-UserBackendFunction swagger-ui generate

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

local-zip: build-builder
	docker create --name $(CONTAINER_NAME) $(BUILDER_NAME) sh -c "apk add --no-cache zip && zip /app/function.zip /app/bootstrap"
	docker cp $(CONTAINER_NAME):/app/function.zip .
	docker rm $(CONTAINER_NAME)

swagger-ui:
	docker run --rm -p 8080:8080 \
	-e SWAGGER_JSON=/docs/openapi.yaml \
	-v $(PWD)/docs:/docs \
	swaggerapi/swagger-ui

generate:
	docker run --rm \
		-v ${PWD}/docs:/docs \
		openapitools/openapi-generator-cli:v7.4.0 \
		generate \
		-i /docs/openapi.yaml \
		-g go-server \
		-o /docs/gen