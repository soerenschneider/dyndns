# dyndns
[![Go Report Card](https://goreportcard.com/badge/github.com/soerenschneider/dyndns)](https://goreportcard.com/report/github.com/soerenschneider/dyndns)
![test-workflow](https://github.com/soerenschneider/dyndns/actions/workflows/test.yaml/badge.svg)
![release-workflow](https://github.com/soerenschneider/dyndns/actions/workflows/release-container.yaml/badge.svg)
![golangci-lint-workflow](https://github.com/soerenschneider/dyndns/actions/workflows/golangci-lint.yaml/badge.svg)

Automatically updates DNS records for hosts that don't have a static IP

## Features

📣 Dynamically updates DNS records to match their public IP address<br/>
🏰 Built-in resiliency for different failure scenarios<br/>
🚀 Dispatch update requests via [AWS SQS](https://aws.amazon.com/sqs/), MQTT and/or HTTP API calls<br/>
🚏 Detect public IP address by network interface's status or by calling HTTP APIs<br/>
🎭 Runs in client / server mode to limit blast-radius of leaked credentials<br/>
🔧 Customizable list of HTTP APIs to detect IP updates<br/>
❤️ Specify HTTP APIs that detect IP updates which are preferred over other API endpoints<br/>
🔐 Messages are cryptographically signed, therefore public MQTT brokers can be used<br/>
🔑 Can use either dynamic credentials using Hashicorp Vault or static credentials<br/>
🔭 Observability through Prometheus metrics

## Why would I need it?

📌 You don't have a static public IP address but want to connect to your home network<br/>
🤹 Ideally, you have multiple endpoints you want to assign DNS records to<br/>

## Installation

### Docker / Podman
```shell
$ docker pull ghcr.io/soerenschneider/dyndns:latest
```

The image is built for `linux/amd64` and `linux/arm64` and signed with cosign:
```shell
$ cosign verify ghcr.io/soerenschneider/dyndns:latest \
    --certificate-identity "https://github.com/soerenschneider/dyndns/.github/workflows/release-container.yaml@refs/heads/main" \
    --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
```

### Binaries
Head over to the [prebuilt binaries](https://github.com/soerenschneider/dyndns/releases) and download the correct binary for your system, together with `checksum.sha256` and its signatures.

```shell
# cosign: made by this repository's release workflow
$ cosign verify-blob checksum.sha256 \
    --bundle checksum.sha256.sigstore.json \
    --certificate-identity "https://github.com/soerenschneider/dyndns/.github/workflows/release-container.yaml@refs/heads/main" \
    --certificate-oidc-issuer "https://token.actions.githubusercontent.com"

# signify, if the release has a checksum.sha256.sig
$ signify -V -p github.pub -m checksum.sha256

# then the binaries themselves
$ sha256sum --ignore-missing -c checksum.sha256
```

### From Source
As a prerequesite, you need to have [Golang SDK](https://go.dev/dl/) installed. After that, you can install dyndns from source by invoking:
```text
$ go install github.com/soerenschneider/dyndns@latest
```

## Configuration

Head over to the [configuration section](docs/configuration.md) to see more details.


## Getting Started

First, you need to build a keypair. This is easily done
```bash
$ docker run ghcr.io/soerenschneider/dyndns-client -gen-keypair
{"public_key":"IyXH8z/+vRsIUEAldlGgKKFcVHoll8w2tzC6o9717m8=","private_key":"h7jrhYupN0LVPnVWqFun6sN+bWNr0B0mh7/mgRaKnhsjJcfzP/69GwhQQCV2UaAooVxUeiWXzDa3MLqj3vXubw=="}
```

# Architecture

## Client Internals
Client states and client failure scenarios and their remediation is found [here](docs/client.md).

## Message format

Data sent over the wire is expected to have the following format, encoded as a JSON message.

### `UpdateRecordRequest` reference
| Field Name  | Description                                 | JSON Key      | Data Type   | Optional |
|-------------|---------------------------------------------|---------------|-------------|----------|
| `PublicIp`  | The resolved IP address.                    | `"public_ip"` | `DnsRecord` | No       |
| `Signature` | The signature associated with the envelope. | `"signature"` | String      | No       |

### `DnsRecord` reference
| Field Name  | Description                                           | JSON Key      | Data Type | Optional |
|-------------|-------------------------------------------------------|---------------|-----------|----------|
| `IpV4`      | The IPv4 address (optional).                          | `"ipv4"`      | String    | Yes      |
| `IpV6`      | The IPv6 address (optional).                          | `"ipv6"`      | String    | Yes      |
| `Host`      | The hostname associated with the resolved IP address. | `"host"`      | String    | No       |
| `Timestamp` | The timestamp when the resolution occurred.           | `"timestamp"` | Time      | No       |


## Observability
Head over to the [metrics](docs/metrics.md) to see more details.

## Releases

release-please opens a release PR on every push to main. Merging it tags the release, and the release workflow builds the binaries and the container image, and signs both with cosign.

Builds are reproducible: the Makefile pins the Go toolchain from `go.mod` and strips paths, VCS data and build IDs, so `make release` on a tag produces the same binaries on any host, and the binary in the image is the released one for its platform. `make signed-release` uses that: it rebuilds a published release, checks that it is byte-identical, and adds a signify signature of its checksums.

## Changelog
The changelog can be found [here](CHANGELOG.md)
