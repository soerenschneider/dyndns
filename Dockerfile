# syntax=docker/dockerfile:1

# Keep in sync with the toolchain in go.mod, otherwise the build downloads that
# toolchain. The build stage runs on the build host and cross-compiles, so no
# emulation is needed.
FROM --platform=$BUILDPLATFORM golang:1.26.4 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT_HASH=unknown
ARG TARGETOS
ARG TARGETARCH

# Same flags as the release binaries, so the binary in the image is identical
# to the released one for its platform.
RUN GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" make compile VERSION="${VERSION}" COMMIT_HASH="${COMMIT_HASH}" OUTPUT=/out/dyndns

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/dyndns /dyndns
ENTRYPOINT ["/dyndns"]
