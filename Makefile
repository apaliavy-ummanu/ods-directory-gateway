ifneq ($(USE_LOCAL_DOCKER),)
    include config/local_docker/.env
else
	include config/local/.env
endif
export

DOCKER_IMAGE ?= ods-directory-gateway
DOCKER_TAG   ?= latest
DOCKER_PORT  ?= 8888

deps:
	go mod download && go mod tidy

build-server: ##@ Builds just the server binary
	go build -o bin/api ./cmd/app

oapi_codegen:
	oapi-codegen -generate skip-prune,types -o client/http/types.gen.go -package http api/http/openapi.yml
	oapi-codegen -generate skip-prune,client -o client/http/client.gen.go -package http api/http/openapi.yml
	oapi-codegen -package http api/http/openapi.yml > internal/service/common/ports/http/http.gen.go

run:
	go run cmd/app/main.go

test-short: ##@ Runs short tests
	go clean -testcache
	go test -coverprofile=coverage.out -short ./...

coverage:
	go tool cover -html=coverage.out

fmt:
	go fmt ./...

check-fmt:
	test -z $$(go fmt ./...)

lint:
	golangci-lint run -c ./config/golangci.yml --timeout=5m

fmt-lint: fmt lint

fakes:
	go generate ./...

.PHONY: docker-build
docker-build:
	docker build -f deployments/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run: docker-build
	docker run --rm -p $(DOCKER_PORT):$(DOCKER_PORT) --env-file config/local_docker/.env \
	$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-shell
docker-shell: docker-build
	docker run --rm -it -p $(DOCKER_PORT):$(DOCKER_PORT) $(DOCKER_IMAGE):$(DOCKER_TAG) /bin/sh