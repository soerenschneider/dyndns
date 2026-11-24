FROM golang:1.26.4 AS builder

ARG VERSION=dev
ARG COMMIT_HASH=unknown
ENV MODULE=github.com/soerenschneider/dyndns/v2

WORKDIR /build/

COPY go.mod go.sum /build/
RUN go mod download

COPY . /build/
# Same flags as the Makefile, so the binary is reproducible.
RUN CGO_ENABLED=0 go build -mod=readonly -trimpath -buildvcs=false \
    -ldflags="-s -w -buildid= -X '$MODULE/internal.BuildVersion=$VERSION' -X '$MODULE/internal.CommitHash=$COMMIT_HASH' -X '$MODULE/internal.GoVersion=$(go env GOVERSION | sed 's/^go//')'" \
    -o dyndns ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /build/dyndns /dyndns
ENTRYPOINT ["/dyndns"]
