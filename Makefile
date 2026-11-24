BUILD_DIR = builds
MODULE = github.com/soerenschneider/dyndns/v2
BINARY_NAME = dyndns
CHECKSUM_FILE = checksum.sha256
SIGNATURE_KEYFILE = ~/.signify/github.sec
DOCKER_PREFIX = ghcr.io/soerenschneider

# Reproducible builds: pin the toolchain from go.mod instead of using whatever go
# is installed, never touch go.mod/go.sum, strip local paths and VCS state (the
# commit is set via ldflags) and build without cgo. The microarchitecture levels
# are set explicitly, so GOAMD64/GOARM64 in the environment cannot change them.
# The same commit then yields bit-identical binaries on every machine.
export GOTOOLCHAIN := $(shell awk '/^toolchain /{print $$2}' go.mod)
export GOFLAGS := -mod=readonly -trimpath -buildvcs=false
export CGO_ENABLED := 0

# Both can be passed in, e.g. by CI or the Dockerfile, which have no tags or no .git.
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
COMMIT_HASH ?= $(shell git rev-parse HEAD)
GO_VERSION := $(GOTOOLCHAIN:go%=%)
LDFLAGS = -s -w -buildid= -X '$(MODULE)/internal.BuildVersion=$(VERSION)' -X '$(MODULE)/internal.CommitHash=$(COMMIT_HASH)' -X '$(MODULE)/internal.GoVersion=$(GO_VERSION)'
GO_BUILD = go build -ldflags="$(LDFLAGS)"

.PHONY: tests clean build compile release checksums sign signed-release check-tree check-tag cross-build docker-build fmt pre-commit-init docs

tests:
	CGO_ENABLED=1 go test -race ./... -cover

clean:
	rm -rf ./$(BUILD_DIR)

build:
	$(GO_BUILD) -o $(BINARY_NAME) ./cmd

# compile builds with the release flags for the Dockerfile. OUTPUT must be passed,
# GOOS and GOARCH set in the environment.
compile:
	GOAMD64=v1 GOARM64=v8.0 $(GO_BUILD) -o $(OUTPUT) ./cmd

cross-build:
	GOOS=linux GOARCH=amd64 GOAMD64=v1   $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-x86_64  ./cmd
	GOOS=linux GOARCH=arm GOARM=5        $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv5   ./cmd
	GOOS=linux GOARCH=arm GOARM=6        $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv6   ./cmd
	GOOS=linux GOARCH=arm64 GOARM64=v8.0 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-aarch64 ./cmd

release: check-tree clean cross-build checksums

# The checksums list bare file names, so "sha256sum -c" works in the download directory.
checksums:
	cd $(BUILD_DIR) && LC_ALL=C sha256sum $(BINARY_NAME)-* > $(CHECKSUM_FILE)

# Signs keyless and writes $(CHECKSUM_FILE).sigstore.json. The signature is only
# trusted when made by the release workflow, so this only runs in CI.
sign:
	@if [ "$$CI" != "true" ]; then \
		echo "keyless signatures are only trusted from the release workflow"; false; \
	fi
	cosign sign-blob --yes --bundle $(BUILD_DIR)/$(CHECKSUM_FILE).sigstore.json $(BUILD_DIR)/$(CHECKSUM_FILE)

# Optional. Run on the release tag after the release workflow has published it:
# rebuilds the release, checks that it is byte-identical to the published one, and
# adds a signify signature of the published checksums. Nothing the workflow uploaded
# is replaced, so its cosign signature stays valid.
signed-release: check-tag release
	gh release download "$(VERSION)" --pattern $(CHECKSUM_FILE) --output $(BUILD_DIR)/$(CHECKSUM_FILE).published --clobber
	cmp $(BUILD_DIR)/$(CHECKSUM_FILE) $(BUILD_DIR)/$(CHECKSUM_FILE).published || \
		{ echo 'Local build differs from the published release'; false; }
	pass keys/signify/github | signify -S -s $(SIGNATURE_KEYFILE) -m $(BUILD_DIR)/$(CHECKSUM_FILE)
	gh release upload "$(VERSION)" $(BUILD_DIR)/$(CHECKSUM_FILE).sig --clobber

check-tree:
	@test -z "$$(git status --porcelain)" || { echo 'Dirty work tree'; false; }

check-tag:
	@git describe --tags --exact-match >/dev/null || { echo 'HEAD is not a release tag'; false; }

docker-build:
	docker build -t "$(DOCKER_PREFIX)/$(BINARY_NAME)" --build-arg VERSION=$(VERSION) --build-arg COMMIT_HASH=$(COMMIT_HASH) .

fmt:
	find . -iname "*.go" -exec go fmt {} \;

pre-commit-init:
	pre-commit install
	pre-commit install --hook-type commit-msg

docs:
	rm -rf go-diagrams
	go run doc/main.go
	cd go-diagrams && dot -Tpng diagram.dot > ../overview.png
