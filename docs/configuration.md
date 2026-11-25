# Configuration

## Client Configuration

| Field           | Type            | JSON Field                   | Environment Variable                |
|-----------------|-----------------|------------------------------|-------------------------------------|
| Host            | string          | host                         | DYNDNS_HOST                         |
| AddrFamilies    | []string        | address_families             | DYNDNS_ADDRESS_FAMILIES             |
| KeyPairPath     | string          | keypair_path                 | DYNDNS_KEYPAIR_PATH                 |
| PreferredUrls   | []string        | http_resolver_preferred_urls | DYNDNS_HTTP_RESOLVER_PREFERRED_URLS |
| FallbackUrls    | []string        | http_resolver_fallback_urls  | DYNDNS_HTTP_RESOLVER_FALLBACK_URLS  |
| NetworkInterface | string         | interface                    | -                                   |
| Ipv4Resolver    | string          | ipv4_resolver                | DYNDNS_IPV4_RESOLVER                |
| Ipv6Resolver    | string          | ipv6_resolver                | DYNDNS_IPV6_RESOLVER                |
| Once            | bool            | -                            | -                                   |
| MqttConfig      | MqttConfig      | -                            | -                                   |
| EmailConfig     | EmailConfig     | notifications                | -                                   |
| InterfaceConfig | InterfaceConfig | -                            | -                                   |

### Resolving IPv4 and IPv6 with different mechanisms

`ipv4_resolver` and `ipv6_resolver` choose how each address family is resolved, either `http` (ask external
services for the public address) or `interface` (read the address from `interface`). Unset, a family uses
`interface` if one is configured and `http` otherwise. A resolver may only be set for a family listed in
`address_families`.

Behind NAT or CGNAT the interface only carries a private IPv4, while the IPv6 on it is globally routable:

```yaml
address_families: [ip4, ip6]
interface: eth0
ipv4_resolver: http
ipv6_resolver: interface
```

Interface-based IPv6 detection is only supported on Linux.

### Addresses that are not publicly routable

The client never publishes private, CGNAT (`100.64.0.0/10`), link-local, documentation or other special-purpose
addresses. When a family resolves to such an address, the client asks the server to delete that family's record,
so it does not keep pointing at a stale address, and publishes whatever usable address is left. If no family has a
usable address, the request only deletes records. This is counted in `dyndns_client_ip_resolves_dropped_total`.
A family that fails to resolve is left untouched; only a definite non-public answer deletes its record.

If the host's own outgoing IPv4 address is a CGNAT address, the http resolver does not ask its providers: they would
only report the carrier's shared address, which does not reach this host. CGNAT on an upstream router (double NAT)
cannot be detected; if that applies, remove `ip4` from `address_families`.

Deleting records requires the `route53:ListResourceRecordSets` permission in addition to
`route53:ChangeResourceRecordSets`.

## MqttConfig

| Field          | Type     | JSON Field      | Environment Variable |
|----------------|----------|-----------------|----------------------|
| Brokers        | []string | brokers         | DYNDNS_BROKERS       |
| ClientId       | string   | client_id       | DYNDNS_CLIENT_ID     |
| CaCertFile     | string   | tls_ca_cert     | DYNDNS_TLS_CA        |
| ClientCertFile | string   | tls_client_cert | DYNDNS_TLS_CERT      |
| ClientKeyFile  | string   | tls_client_key  | DYNDNS_TLS_KEY       |
| TlsInsecure    | bool     | tls_insecure    | DYNDNS_TLS_INSECURE  |


## EmailConfig

| Field        | Type     | JSON Field | Environment Variable  |
|--------------|----------|------------|-----------------------|
| From         | string   | from       | DYNDNS_EMAIL_FROM     |
| To           | []string | to         | DYNDNS_EMAIL_TO       |
| SmtpHost     | string   | host       | DYNDNS_EMAIL_HOST     |
| SmtpPort     | int      | port       | DYNDNS_EMAIL_PORT     |
| SmtpUsername | string   | user       | DYNDNS_EMAIL_USER     |
| SmtpPassword | string   | password   | DYNDNS_EMAIL_PASSWORD |


## Server Config

| Field           | Type                | JSON Field     | Environment Variable |
|-----------------|---------------------|----------------|----------------------|
| KnownHosts      | map[string][]string | known_hosts    | -                    |
| HostedZoneId    | string              | hosted_zone_id | -                    |
| MqttConfig      | MqttConfig          | -              | -                    |
| Aws             | AwsCredentials      | aws            | DYNDNS_AWS_*         |
| EmailConfig     | EmailConfig         | notifications  | -                    |


## AWS Credentials

The server needs AWS credentials for Route 53 and, if it listens on SQS, for SQS. A client needs them to send updates via SQS.
Both use the AWS SDK's default credential chain: the `AWS_*` environment variables, the shared config and credentials
files, and the role of an EC2 instance, ECS task or EKS pod. To use the access key of an IAM user instead, e.g. one
created by `deployment/terraform/sqs`, set it in the `aws` section of `client` or `server`:

| Field               | Type   | JSON Field             | Environment Variable                 |
|---------------------|--------|------------------------|--------------------------------------|
| AccessKeyId         | string | access_key_id          | DYNDNS_AWS_ACCESS_KEY_ID             |
| SecretAccessKey     | string | secret_access_key      | DYNDNS_AWS_SECRET_ACCESS_KEY         |
| SecretAccessKeyFile | string | secret_access_key_file | DYNDNS_AWS_SECRET_ACCESS_KEY_FILE    |

`access_key_id` requires exactly one of `secret_access_key` and `secret_access_key_file`.

```yaml
server:
  aws:
    access_key_id: AKIA...
    secret_access_key_file: /run/secrets/aws-secret-access-key
```

## Reference
| Keyword        | Description                                    | Example                      | Mandatory |
|----------------|------------------------------------------------|------------------------------|-----------|
| host           | FQDN of the host you want to update the IP for | home.example.com             | Y         |
| keypair_path   | The path to the ed25519 keypair                | /etc/dyndns/keypair          | Y         |
| metrics_listen | HTTP metrics handler listen address            | :9095                        | N         |
| brokers        | The MQTT brokers to connect to                 | ["tcp://host.tld":1883]      | Y         |
| client_id      | Client id for the MQTT connection              | crazy-horse                  | Y         |
| interface      | The network interface to check for IP updates  | eth0                         | N         |
