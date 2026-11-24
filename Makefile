BUILD_DIR = builds
MODULE = github.com/soerenschneider/dyndns/v2
BINARY_NAME = dyndns
CHECKSUM_FILE = $(BUILD_DIR)/checksum.sha256
SIGNATURE_KEYFILE = ~/.signify/github.sec
DOCKER_PREFIX = ghcr.io/soerenschneider

# Reproducible builds: pin the toolchain from go.mod instead of using whatever go
# is installed, never touch go.mod/go.sum, strip local paths and VCS state (the
# commit is set via ldflags) and build without cgo.
# The same commit then yields bit-identical binaries on every machine.
export GOTOOLCHAIN := $(shell awk '/^toolchain /{print $$2}' go.mod)
export GOFLAGS := -mod=readonly -trimpath -buildvcs=false
export CGO_ENABLED := 0

VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
COMMIT_HASH := $(shell git rev-parse HEAD)
GO_VERSION := $(GOTOOLCHAIN:go%=%)
LDFLAGS = -s -w -buildid= -X '$(MODULE)/internal.BuildVersion=$(VERSION)' -X '$(MODULE)/internal.CommitHash=$(COMMIT_HASH)' -X '$(MODULE)/internal.GoVersion=$(GO_VERSION)'

.PHONY: tests clean build release signed-release lambda-server cross-build docker-build fmt pre-commit-init docs

tests:
	CGO_ENABLED=1 go test -race ./... -cover

clean:
	rm -rf ./$(BUILD_DIR)

build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) ./cmd

release: clean cross-build
	sha256sum $(BUILD_DIR)/$(BINARY_NAME)-* > $(CHECKSUM_FILE)

signed-release: release
	pass keys/signify/github | signify -S -s $(SIGNATURE_KEYFILE) -m $(CHECKSUM_FILE)
	gh-upload-assets -o soerenschneider -r dyndns -f ~/.gh-token builds

lambda-server:
	GOOS=linux GOARCH=arm64 go build -tags server,aws -ldflags="$(LDFLAGS)" -o bootstrap cmd/server/server_lambda.go
	rm -f dyndns-server-lambda.zip
	zip dyndns-server-lambda.zip bootstrap
	rm -f bootstrap

cross-build:
	GOOS=linux GOARCH=amd64       go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-x86_64  ./cmd
	GOOS=linux GOARCH=arm GOARM=5 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv5   ./cmd
	GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv6   ./cmd
	GOOS=linux GOARCH=arm64       go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-aarch64 ./cmd

docker-build:
	docker build -t "$(DOCKER_PREFIX)/$(BINARY_NAME)" .

fmt:
	find . -iname "*.go" -exec go fmt {} \;

pre-commit-init:
	pre-commit install
	pre-commit install --hook-type commit-msg

docs:
	rm -rf go-diagrams
	go run doc/main.go
	cd go-diagrams && dot -Tpng diagram.dot > ../overview.png
