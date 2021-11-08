ARG MODE=dyndns-client

FROM golang:1.17.2 as builder
ARG MODE
ENV MODE="$MODE"
ENV MODULE=github.com/soerenschneider/dyndns
WORKDIR /build/
ADD . /build/
RUN go build -ldflags="-X $MODULE/internal.BuildVersion=$(git describe --tags --abbrev=0 || echo dev) -X $MODULE/internal.CommitHash=$(git rev-parse HEAD)" -o "dyndns-$MODE" "cmd/$MODE/$MODE.go"

FROM gcr.io/distroless/base
ARG MODE
ENV MODE="$MODE"
COPY --from=builder "/build/dyndns-$MODE" /dyndns
ENTRYPOINT ["/dyndns"]