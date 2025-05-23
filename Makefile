BUILD_DIR = builds
MODULE = github.com/soerenschneider/dyndns
BINARY_NAME_SERVER = dyndns-server
BINARY_NAME_CLIENT = dyndns-client
CHECKSUM_FILE = $(BUILD_DIR)/checksum.sha256
SIGNATURE_KEYFILE = ~/.signify/github.sec
DOCKER_PREFIX = ghcr.io/soerenschneider

tests:
	go test ./... -tags client,server -cover

clean:
	rm -rf ./$(BUILD_DIR)

build: version-info
	go build -race -ldflags="-w -X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}' -X '$(MODULE)/internal.GoVersion=${GO_VERSION}'" -o $(BINARY_NAME_SERVER) -tags server cmd/server/server.go
	go build -race -ldflags="-w -X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}' -X '$(MODULE)/internal.GoVersion=${GO_VERSION}'" -o $(BINARY_NAME_CLIENT) -tags client cmd/client/client.go

release: clean version-info cross-build-client cross-build-server
	sha256sum $(BUILD_DIR)/dyndns-* > $(CHECKSUM_FILE)

signed-release: release
	pass keys/signify/github | signify -S -s $(SIGNATURE_KEYFILE) -m $(CHECKSUM_FILE)
	gh-upload-assets -o soerenschneider -r dyndns -f ~/.gh-token builds

lambda-server:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags server,aws -ldflags="-w -X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}' -X '$(MODULE)/internal.GoVersion=${GO_VERSION}'" -o bootstrap cmd/server/server_lambda.go
	rm -f dyndns-server-lambda.zip
	zip dyndns-server-lambda.zip bootstrap
	rm -f bootstrap

cross-build-server:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0       go build -ldflags="-w -X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}' -X '$(MODULE)/internal.GoVersion=${GO_VERSION}'" -o $(BUILD_DIR)/$(BINARY_NAME_SERVER)-linux-x86_64    -tags server cmd/server/server.go
	GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=0 go build -ldflags="-w -X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}' -X '$(MODULE)/internal.GoVersion=${GO_VERSION}'" -o $(BUILD_DIR)/$(BINARY_NAME_SERVER)-linux-armv5     -tags server cmd/server/server.go
	GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build -ldflags="-w -X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}' -X '$(MODULE)/internal.GoVersion=${GO_VERSION}'" -o $(BUILD_DIR)/$(BINARY_NAME_SERVER)-linux-armv6     -tags server cmd/server/server.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0       go build -ldflags="-w -X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}' -X '$(MODULE)/internal.GoVersion=${GO_VERSION}''" -o $(BUILD_DIR)/$(BINARY_NAME_SERVER)-linux-aarch64   -tags server cmd/server/server.go

cross-build-client:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0       go build -ldflags="-w -X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}' -X '$(MODULE)/internal.GoVersion=${GO_VERSION}'" -o $(BUILD_DIR)/$(BINARY_NAME_CLIENT)-linux-x86_64    -tags client cmd/client/client.go
	GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=0 go build -ldflags="-w -X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}' -X '$(MODULE)/internal.GoVersion=${GO_VERSION}'" -o $(BUILD_DIR)/$(BINARY_NAME_CLIENT)-linux-armv5     -tags client cmd/client/client.go
	GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build -ldflags="-w -X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}' -X '$(MODULE)/internal.GoVersion=${GO_VERSION}'" -o $(BUILD_DIR)/$(BINARY_NAME_CLIENT)-linux-armv6     -tags client cmd/client/client.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0       go build -ldflags="-w -X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}' -X '$(MODULE)/internal.GoVersion=${GO_VERSION}'" -o $(BUILD_DIR)/$(BINARY_NAME_CLIENT)-linux-aarch64   -tags client cmd/client/client.go

docker-build-server:
	docker build -t "$(DOCKER_PREFIX)/$(BINARY_NAME_SERVER)" --build-arg MODE=server .

docker-build-client:
	docker build -t "$(DOCKER_PREFIX)/$(BINARY_NAME_CLIENT)" --build-arg MODE=client .

docker-build: docker-build-server docker-build-client

version-info:
	$(eval VERSION := $(shell git describe --tags --abbrev=0 || echo "dev"))
	$(eval COMMIT_HASH := $(shell git rev-parse HEAD))
	$(eval GO_VERSION := $(shell go version | awk '{print $$3}' | sed 's/^go//'))

fmt:
	find . -iname "*.go" -exec go fmt {} \; 

pre-commit-init:
	pre-commit install
	pre-commit install --hook-type commit-msg

docs:
	rm -rf go-diagrams
	go run doc/main.go
	cd go-diagrams && dot -Tpng diagram.dot > ../overview.png
