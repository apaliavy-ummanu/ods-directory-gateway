deps:
	go mod download && go mod tidy


oapi_codegen:
	oapi-codegen -generate skip-prune,types -o client/http/types.gen.go -package http api/http/openapi.yml
	oapi-codegen -generate skip-prune,client -o client/http/client.gen.go -package http api/http/openapi.yml
	oapi-codegen -package http api/http/openapi.yml > internal/service/common/ports/http/http.gen.go

run:
	go run cmd/app/main.go