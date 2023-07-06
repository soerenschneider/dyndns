[![Go Report Card](https://goreportcard.com/badge/github.com/soerenschneider/dyndns)](https://goreportcard.com/report/github.com/soerenschneider/dyndns)
# dyndns

Automatically updates DNS records for internet connections that don't have a static IP

## Goals

### Resilience
#### Multiple Brokers
It's possible to specify multiple MQTT brokers at once. Instead of only connecting to one broker at a time, dyndns 
tries to be connected to all configured brokers all the time.

#### Multiple Servers
Dyndns allows multiple servers to listen simultaneously for requests and try to upsert them. Instead of using complicated consensus mechanism, upserts for host records are expected to be idempotent.

#### Detection of IP updates
Currently, two methods are support to detect IP updates
1. Checking the IP address of the network interface connected to the internet. While this is very reliable, the client needs to run on your router.
2. Using external websites, such as ipinfo.net, that return your ip address. Trust is shifted to the website providing your IP address but it can be run on any computer within your network.

### Security
#### Pre-defined hostnames
Hosts have to be pre-defined in the dyndns server's configuration before requests are accepted for that host. This is done by specifying (multiple) [ed25519](https://ed25519.cr.yp.to/) [keypairs](https://en.wikipedia.org/wiki/Public-key_cryptography).

#### Spoofing protection
Update request payloads are signed using the host's configured private key. The signature is verified by the server component.

#### Replay attacks
A (signed) timestamp is included in the update request payload so even when you're using a public MQTT broker, replayed messages won't lead to updated host records.

#### Least privilege
The clients run without access to credentials of the DNS provider. The worst case scenario is the ability to change a single host record. Obviously this only makes sense if you have more than one host record that should be synchronized.

The dyndns server should only run a trusted and well-secured server though.

#### Support for dynamic IAM credentials
Dyndns supports acquiring dynamic IAM credentials through Vault's [AWS secret engine](https://developer.hashicorp.com/vault/docs/secrets/aws)

### Observability

#### Metrics
Prometheus metrics are supported and documented below. It's easy to define dashboards and alerts to make sure all is up and well.

#### Notifications
Notifications (such as emails) can be used to display events such as detection of updated IPs and applied updates.


(To reach some goals outlined above) the high level architecture is as follows.
# Architecture

## Message format
```go
type Envelope struct {
	PublicIp  ResolvedIp `json:"public_ip"`
	Signature string     `json:"signature"`
}

type ResolvedIp struct {
	IpV4      string    `json:"ipv4,omitempty"`
	IpV6      string    `json:"ipv6,omitempty"`
	Host      string    `json:"host"`
	Timestamp time.Time `json:"timestamp"`
}
```


### Reference

### Client Configuration

| Field           | Type            | JSON Field                   | Environment Variable                |
|-----------------|-----------------|------------------------------|-------------------------------------|
| Host            | string          | host                         | DYNDNS_HOST                         |
| AddrFamilies    | []string        | address_families             | DYNDNS_ADDRESS_FAMILIES             |
| KeyPairPath     | string          | keypair_path                 | DYNDNS_KEYPAIR_PATH                 |
| MetricsListener | string          | metrics_listen               | DYNDNS_METRICS_LISTEN               |
| PreferredUrls   | []string        | http_resolver_preferred_urls | DYNDNS_HTTP_RESOLVER_PREFERRED_URLS |
| FallbackUrls    | []string        | http_resolver_fallback_urls  | DYNDNS_HTTP_RESOLVER_FALLBACK_URLS  |
| Once            | bool            | -                            | -                                   |
| MqttConfig      | MqttConfig      | -                            | -                                   |
| EmailConfig     | EmailConfig     | notifications                | -                                   |
| InterfaceConfig | InterfaceConfig | -                            | -                                   |

### MqttConfig

| Field          | Type     | JSON Field      | Environment Variable |
|----------------|----------|-----------------|----------------------|
| Brokers        | []string | brokers         | DYNDNS_BROKERS       |
| ClientId       | string   | client_id       | DYNDNS_CLIENT_ID     |
| CaCertFile     | string   | tls_ca_cert     | DYNDNS_TLS_CA        |
| ClientCertFile | string   | tls_client_cert | DYNDNS_TLS_CERT      |
| ClientKeyFile  | string   | tls_client_key  | DYNDNS_TLS_KEY       |
| TlsInsecure    | bool     | tls_insecure    | DYNDNS_TLS_INSECURE  |


### EmailConfig

| Field        | Type     | JSON Field | Environment Variable  |
|--------------|----------|------------|-----------------------|
| From         | string   | from       | DYNDNS_EMAIL_FROM     |
| To           | []string | to         | DYNDNS_EMAIL_TO       |
| SmtpHost     | string   | host       | DYNDNS_EMAIL_HOST     |
| SmtpPort     | int      | port       | DYNDNS_EMAIL_PORT     |
| SmtpUsername | string   | user       | DYNDNS_EMAIL_USER     |
| SmtpPassword | string   | password   | DYNDNS_EMAIL_PASSWORD |


### Server Config

| Field           | Type                | JSON Field     | Environment Variable |
|-----------------|---------------------|----------------|----------------------|
| KnownHosts      | map[string][]string | known_hosts    | -                    |
| HostedZoneId    | string              | hosted_zone_id | -                    |
| MetricsListener | string              | metrics_listen | -                    |
| MqttConfig      | MqttConfig          | -              | -                    |
| VaultConfig     | VaultConfig         | -              | -                    |
| EmailConfig     | EmailConfig         | notifications  | -                    |


### Vault Config
Here's a markdown table that displays the name, type, JSON field name, and environment variable name (if applicable) for each field in the `VaultConfig` struct:

| Field           | Type              | JSON Field            | Environment Variable |
|-----------------|-------------------|-----------------------|----------------------|
| VaultAddr       | string            | vault_addr            | -                    |
| AuthStrategy    | VaultAuthStrategy | vault_auth_strategy   | -                    |
| AwsRoleName     | string            | vault_aws_role_name   | -                    |
| AwsMountPath    | string            | vault_aws_mount_path  | -                    |
| AppRoleId       | string            | vault_app_role_id     | -                    |
| AppRoleSecretId | string            | vault_app_role_secret | -                    |
| VaultToken      | string            | vault_token           | -                    |

Please note that some fields do not have corresponding environment variable names as they are not specified in the `env` tag.

### Reference
| Keyword        | Description                                    | Example             | Mandatory |
|----------------|------------------------------------------------|---------------------|-----------|
| host           | FQDN of the host you want to update the IP for | https://vault:8200  | Y         |
| keypair_path   | The path to the ed25519 keypair                | /etc/dyndns/keypair | Y         |
| metrics_listen | HTTP metrics handler listen address            | :9095               | N         |
| brokers        | The MQTT brokers to connect to                 | ["tcp://host.tld":1883]      | Y         |
| client_id      | Client id for the MQTT connection              | crazy-horse         | Y         |
| interface      | The network interface to check for IP updates  | eth0                | N         |

# Available Metrics

## Common Metrics (shared in server and client mode)
Here's the markdown table for the updated variables:

| Metric Name                        | Description                                     | Labels                   |
| ---------------------------------- | ----------------------------------------------- | ------------------------ |
| version                            | Version metric                                  | version, hash            |
| start_time_seconds                 | Process start time in seconds                   | N/A                      |
| mqtt_brokers_configured_total      | Total count of configured MQTT brokers          | N/A                      |
| mqtt_brokers_connected_total       | Total count of connected MQTT brokers           | N/A                      |
| mqtt_connections_lost_total        | Total count of lost MQTT connections            | N/A                      |
| notification_errors                | Gauge indicating the number of notification errors | N/A                      |


## Server Metrics
| Metric Name                               | Description                                            | Labels                       |
| ----------------------------------------- | ------------------------------------------------------ | ---------------------------- |
| dyndns_heartbeat_timestamp_seconds        | Server heartbeat timestamp                             | N/A                          |
| dyndns_known_hosts_configuration_hash     | Known hosts configuration hash                        | N/A                          |
| dyndns_dns_propagation_requests_total     | Total count of DNS propagation requests                | N/A                          |
| dyndns_dns_propagation_request_timestamp_seconds | Timestamp of the latest DNS propagation request  | N/A                          |
| dyndns_dns_propagations_total             | Total count of successful DNS propagations             | host                         |
| dyndns_dns_propagations_errors_total      | Total count of DNS propagation errors                  | host                         |
| dyndns_messages_received_total            | Total count of received messages                       | N/A                          |
| dyndns_signature_verifications_errors_total | Total count of signature verification errors         | host                         |
| dyndns_public_keys_missing_total          | Total count of missing public keys                      | host                         |
| dyndns_messages_ignored_total              | Total count of ignored messages                         | host, reason                 |
| dyndns_message_validations_failed_total    | Total count of failed message validations              | host, reason                 |
| dyndns_vault_token_expiry_time_seconds    | Expiry time of the Vault token                          | N/A                          |
| dyndns_config_public_key_errors_total     | Total count of public key configuration errors          | N/A                          |
| dyndns_message_parsing_failed_total       | Total count of failed message parsing                   | N/A                          |


## Client Metrics

| Metric Name                                | Description                                            | Labels                            |
| ------------------------------------------ | ------------------------------------------------------ | --------------------------------- |
| dyndns_ip_resolves_errors_total             | Total count of IP resolve errors                       | host, resolver, name              |
| dyndns_ip_resolved_successful_total         | Total count of successful IP resolves                  | host, resolver, name              |
| dyndns_reconcilers_pending_changes_total    | Total count of pending changes for reconcilers          | host                              |
| dyndns_reconciler_timestamp_seconds         | Timestamp of reconciler activity                       | host                              |
| dyndns_ip_resolves_invalid_total            | Total count of invalid IP resolves                     | host, resolver, url               |
| dyndns_ip_resolves_success_total            | Total count of successful IP resolves                  | host, resolver                    |
| dyndns_ip_resolves_last_check_timestamp_seconds | Timestamp of the last IP resolve check          | host, resolver                    |
| dyndns_updates_dispatch_errors_total        | Total count of update dispatch errors                  | host                              |
| dyndns_updates_dispatched_total             | Total count of dispatched updates                      | N/A                               |
| dyndns_state_changed_timestamp              | Timestamp of state change                              | host, from, to                    |
| dyndns_current_state_bool                   | Current state as a boolean value                       | host, state                       |
| dyndns_resolver_response_time_seconds       | Histogram of resolver response time in seconds         | resolver                          |
