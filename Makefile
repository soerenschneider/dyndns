build-client: version-info
	env CGO_ENABLED=0 go build -ldflags="-X 'dyndns/internal.BuildTime=${BUILD_TIME}' -X 'dyndns/internal.BuildVersion=${VERSION}' -X 'dyndns/internal.CommitHash=${COMMIT_HASH}'" -o dyndns-client cmd/client/client.go

build-server: version-info
	env CGO_ENABLED=0 go build -ldflags="-X 'dyndns/internal.BuildTime=${BUILD_TIME}' -X 'dyndns/internal.BuildVersion=${VERSION}' -X 'dyndns/internal.CommitHash=${COMMIT_HASH}'" -o dyndns-server cmd/server/server.go

build: build-client build-server

tests:
	go test ./...

version-info:
	$(eval VERSION := $(shell git describe --tags || echo "dev"))
	$(eval BUILD_TIME := $(shell date --rfc-3339=seconds))
	$(eval COMMIT_HASH := $(shell git rev-parse --short HEAD))
